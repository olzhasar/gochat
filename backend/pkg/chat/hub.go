package chat

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/olzhasar/gochat/pkg/metrics"
	"github.com/olzhasar/gochat/pkg/protocol"
)

const EMPTY_ROOM_TIMEOUT = 1 * time.Minute

type Hub struct {
	rooms        map[string]*Room
	roomsLock   sync.RWMutex
	clients      map[string]*Client
	clientsLock sync.Mutex
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

	h.roomsLock.Lock()
	for {
		id := generateID()
		if _, ok := h.rooms[id]; !ok {
			room = newRoom(id)
			h.rooms[id] = &room
			break
		}
	}
	h.roomsLock.Unlock()

	room.run()

	metrics.RoomCount.Inc()

	return &room
}

func (h *Hub) CreateClient() *Client {
	var client Client

	h.clientsLock.Lock()
	for {
		id := generateID()
		if _, ok := h.rooms[id]; !ok {
			client = newClient(id)
			h.clients[id] = &client
			break
		}
	}
	h.clientsLock.Unlock()

	metrics.ClientCount.Inc()

	return &client
}

func (h *Hub) GetRoom(id string) *Room {
	h.roomsLock.RLock()
	defer h.roomsLock.RUnlock()
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	h.roomsLock.RLock()
	defer h.roomsLock.RUnlock()
	return len(h.rooms)
}

func (h *Hub) TerminateClient(client *Client) {
	if client.room != nil {
		client.room.leave(client)
	}
	client.close()

	h.clientsLock.Lock()
	delete(h.clients, client.id)
	h.clientsLock.Unlock()
}

func (h *Hub) Dispatch(client *Client, msg *protocol.Message) error {
	room := h.GetRoom(client.room.ID) // FIXME: should read from msg
	if room == nil {
		return errors.New("Room not found")
	}

	return client.receive(room, msg)
}

func generateID() string {
	return uuid.New().String()
}
