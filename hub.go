package main

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const EMPTY_ROOM_TIMEOUT = 5 * time.Minute

type Message struct {
	room    *Room
	author  *Client
	content []byte
}

func NewMessage(author *Client, room *Room, content []byte) Message {
	return Message{author: author, room: room, content: content}
}

type Room struct {
	ID      string
	clients []*Client
}

func (r *Room) ClientCount() int {
	return len(r.clients)
}

type Instruction struct {
	client *Client
	room   *Room
}

func NewInstruction(client *Client, room *Room) Instruction {
	return Instruction{client: client, room: room}
}

type Hub struct {
	registerChan   chan Instruction
	unregisterChan chan Instruction
	broadcastChan  chan Message
	rooms          map[string]*Room
}

func NewHub() *Hub {
	return &Hub{
		registerChan:   make(chan Instruction),
		unregisterChan: make(chan Instruction),
		broadcastChan:  make(chan Message),
		rooms:          make(map[string]*Room),
	}
}

func (h *Hub) register(client *Client, room *Room) {
	h.registerChan <- NewInstruction(client, room)
}

func (h *Hub) unregister(client *Client, room *Room) {
	h.unregisterChan <- NewInstruction(client, room)
}

func (h *Hub) broadcast(message Message) {
	h.broadcastChan <- message
}

func (h *Hub) handleRegister(client *Client, room *Room) {
	room.clients = append(room.clients, client)
	client.listen()
	log.Println("client connected")
	log.Println("clients in the room: ", len(room.clients))
}

func (h *Hub) handleUnregister(client *Client, room *Room) {
	for i, c := range room.clients {
		if c == client {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}

	client.close()

	if room.ClientCount() == 0 {
		h.terminateRoomIfEmpty(room)
	}

	log.Println("client disconnected")
	log.Println("clients in the room: ", len(room.clients))
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

	for _, client := range message.room.clients {
		if client != message.author {
			client.write(encoded)
		}
	}
}

func (h *Hub) run() {
	go func() {
		for {
			select {
			case instruction := <-h.registerChan:
				h.handleRegister(instruction.client, instruction.room)
			case instruction := <-h.unregisterChan:
				h.handleUnregister(instruction.client, instruction.room)
			case message := <-h.broadcastChan:
				h.handleBroadcast(message)
			}
		}
	}()
}

func (h *Hub) listenClient(client *Client, room *Room) {
	for {
		messageType, message, err := client.conn.ReadMessage()
		if err != nil {
			h.unregister(client, room)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		msg := NewMessage(client, room, message)
		h.broadcast(msg)
	}
}

func generateID() string {
	return uuid.New().String()
}

func (h *Hub) CreateRoom() *Room {
	var room *Room

	for {
		id := generateID()
		if h.rooms[id] == nil {
			room = &Room{ID: id}
			h.rooms[room.ID] = room
			break
		}
	}

	go func() {
		time.Sleep(EMPTY_ROOM_TIMEOUT)
		h.terminateRoomIfEmpty(room)
	}()

	return room
}

func (h *Hub) GetRoom(id string) *Room {
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	return len(h.rooms)
}

func (h *Hub) terminateRoomIfEmpty(room *Room) {
	if room.ClientCount() > 0 {
		return
	}
	log.Println("terminating room", room.ID)
	delete(h.rooms, room.ID)
}
