package chat

import (
	"slices"
	"sync"
)

type Room struct {
	ID              string
	clients         []*Client
	clients_lock    sync.RWMutex
	broadcast_queue chan []byte
}

func NewRoom(ID string) Room {
	return Room{
		ID:              ID,
		clients:         make([]*Client, 0),
		broadcast_queue: make(chan []byte),
	}
}

func (r *Room) ClientCount() int {
	r.clients_lock.RLock()
	defer r.clients_lock.RUnlock()
	return len(r.clients)
}

func (r *Room) join(client *Client) {
	r.clients_lock.Lock()
	r.clients = append(r.clients, client)
	r.clients_lock.Unlock()
	// TODO: announce joining
}

func (r *Room) leave(client *Client) {
	r.clients_lock.Lock()
	for i, c := range r.clients {
		if c == client {
			r.clients = slices.Concat(r.clients[:i], r.clients[i+1:])
			break
		}
	}
	r.clients_lock.Unlock()

	if client.name != "" {
		leaveMsg := NewMessage(client, r, MESSAGE_TYPE_LEAVE, nil)
		r.Broadcast(leaveMsg.Encode())
	}
}

// broadcast to all clients inside the room, does not block
func (r *Room) Broadcast(msg []byte) {
	if msg == nil {
		panic("msg nil")
	}
	r.broadcast_queue <- msg
}

func (r *Room) Run() {
	go func() {
		for msg := range r.broadcast_queue {
			for _, client := range r.clients {
				client.Send(msg)
			}
		}
	}()
}
