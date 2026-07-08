package chat

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/olzhasar/gochat/pkg/metrics"
)

const EMPTY_ROOM_TIMEOUT = 1 * time.Minute

type Hub struct {
	rooms        map[string]*Room
	rooms_lock   sync.RWMutex
	clients      map[string]*Client
	clients_lock sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:   make(map[string]*Room),
		clients: make(map[string]*Client),
	}
}

// Creates room and starts the room's listening goroutine
func (h *Hub) CreateRoom() *Room {
	var room Room

	h.rooms_lock.Lock()
	for {
		id := generateID()
		if _, ok := h.rooms[id]; !ok {
			room = newRoom(id)
			h.rooms[id] = &room
			break
		}
	}
	h.rooms_lock.Unlock()

	room.run()

	metrics.RoomCount.Inc()

	return &room
}

func (h *Hub) CreateClient(conn *websocket.Conn) *Client {
	var client Client

	h.clients_lock.Lock()
	for {
		id := generateID()
		if _, ok := h.rooms[id]; !ok {
			client = newClient(id, conn)
			h.clients[id] = &client
			break
		}
	}
	h.clients_lock.Unlock()

	client.run()

	metrics.ClientCount.Inc()

	return &client
}

func (h *Hub) GetRoom(id string) *Room {
	h.rooms_lock.RLock()
	defer h.rooms_lock.RUnlock()
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	h.rooms_lock.RLock()
	defer h.rooms_lock.RUnlock()
	return len(h.rooms)
}

func generateID() string {
	return uuid.New().String()
}
