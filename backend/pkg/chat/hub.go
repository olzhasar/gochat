package chat

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/olzhasar/gochat/pkg/metrics"
)

const EMPTY_ROOM_TIMEOUT = 1 * time.Minute

type Hub struct {
	rooms     map[string]*Room
	rooms_mut sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

// Creates room and starts the room's listening goroutine
func (h *Hub) CreateRoom() *Room {
	var room Room

	h.rooms_mut.Lock()
	for {
		id := generateID()
		if h.rooms[id] == nil {
			room = NewRoom(id)
			h.rooms[id] = &room
			break
		}
	}
	h.rooms_mut.Unlock()

	room.Run()

	metrics.RoomCount.Inc()

	return &room
}

func (h *Hub) GetRoom(id string) *Room {
	h.rooms_mut.RLock()
	defer h.rooms_mut.RUnlock()
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	h.rooms_mut.RLock()
	defer h.rooms_mut.RUnlock()
	return len(h.rooms)
}

func generateID() string {
	return uuid.New().String()
}
