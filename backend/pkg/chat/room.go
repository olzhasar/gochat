package chat

import (
	"slices"
	"sync"

	"github.com/olzhasar/gochat/pkg/protocol"
)

type Room struct {
	ID          string
	clients     []*Client
	clientsLock sync.RWMutex
	broadcastQ  chan protocol.Message
}

func newRoom(ID string) Room {
	return Room{
		ID:         ID,
		clients:    make([]*Client, 0),
		broadcastQ: make(chan protocol.Message),
	}
}

func (r *Room) ClientCount() int {
	r.clientsLock.RLock()
	defer r.clientsLock.RUnlock()
	return len(r.clients)
}

func (r *Room) join(client *Client) {
	r.clientsLock.Lock()
	r.clients = append(r.clients, client)
	r.clientsLock.Unlock()
	// TODO: announce joining
}

func (r *Room) leave(client *Client) {
	r.clientsLock.Lock()
	for i, c := range r.clients {
		if c == client {
			r.clients = slices.Concat(r.clients[:i], r.clients[i+1:])
			break
		}
	}
	r.clientsLock.Unlock()

	if client.name != "" {
		leaveMsg := protocol.Message{Type: protocol.MessageTypeLeave, ClientID: client.id, RoomID: r.ID}
		r.broadcast(leaveMsg)
	}
}

// broadcast to all clients inside the room, does not block
func (r *Room) broadcast(payload protocol.Message) {
	r.broadcastQ <- payload
}

func (r *Room) run() {
	go func() {
		for msg := range r.broadcastQ {
			r.clientsLock.RLock()
			for _, client := range r.clients {
				client.enqueue(msg)
			}
			r.clientsLock.RUnlock()
		}
	}()
}
