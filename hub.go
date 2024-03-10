package main

import (
	"github.com/google/uuid"
	"log"
)

type Message struct {
	author  *Client
	content []byte
}

type Room struct {
	ID      string
	clients []*Client
}

type Hub struct {
	clients        []*Client
	registerChan   chan *Client
	unregisterChan chan *Client
	broadcastChan  chan Message
	rooms          map[string]*Room
}

func NewHub() *Hub {
	return &Hub{
		registerChan:   make(chan *Client),
		unregisterChan: make(chan *Client),
		broadcastChan:  make(chan Message),
		clients:        make([]*Client, 0),
		rooms:          make(map[string]*Room),
	}
}

func (h *Hub) register(client *Client) {
	h.registerChan <- client
	client.listen()
	h.listenClient(client)
}

func (h *Hub) unregister(client *Client) {
	h.unregisterChan <- client
	client.close()
}

func (h *Hub) broadcast(message Message) {
	h.broadcastChan <- message
}

func (h *Hub) handleRegister(client *Client) {
	h.clients = append(h.clients, client)
	log.Println("client connected")
	log.Println("clients connected: ", len(h.clients))
}

func (h *Hub) handleUnregister(client *Client) {
	for i, c := range h.clients {
		if c == client {
			h.clients = append(h.clients[:i], h.clients[i+1:]...)
			break
		}
	}
	log.Println("client disconnected")
	log.Println("clients connected: ", len(h.clients))
}

func encodeMessage(message Message) []byte {
	prefix := []byte(message.author.name + "|")
	return append(prefix, message.content...)
}

func (h *Hub) handleBroadcast(message Message) {
	if message.author.name == "" {
		message.author.setName(string(message.content))
		log.Println("client set name to", message.author.name)
		return
	}

	log.Println("broadcasting message from", message.author.name)

	encoded := encodeMessage(message)

	for _, client := range h.clients {
		if client != message.author {
			client.write(encoded)
		}
	}
}

func (h *Hub) run() {
	go func() {
		for {
			select {
			case client := <-h.registerChan:
				h.handleRegister(client)
			case client := <-h.unregisterChan:
				h.handleUnregister(client)
			case message := <-h.broadcastChan:
				h.handleBroadcast(message)
			}
		}
	}()
}

func (h *Hub) listenClient(client *Client) {
	go func() {
		for {
			_, message, err := client.conn.ReadMessage()
			if err != nil {
				h.unregister(client)
				return
			}

			msg := Message{author: client, content: message}
			h.broadcast(msg)
		}
	}()
}

func generateID() string {
	return uuid.New().String()
}

func (h *Hub) CreateRoom() *Room {
	for {
		id := generateID()
		if h.rooms[id] == nil {
			room := &Room{ID: id}
			h.rooms[room.ID] = room
			return room
		}
	}
}

func (h *Hub) GetRoom(id string) *Room {
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	return len(h.rooms)
}
