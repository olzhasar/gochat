package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const EMPTY_ROOM_TIMEOUT = 5 * time.Minute

const MESSAGE_TYPE_TEXT = 1
const MESSAGE_TYPE_NAME = 2
const MESSAGE_TYPE_LEAVE = 3
const MESSAGE_TYPE_TYPING = 4
const MESSAGE_TYPE_STOP_TYPING = 5

type Message struct {
	msgType int
	room    *Room
	author  *Client
	content []byte
}

func (m Message) encode() []byte {
	var content string
	if m.msgType == MESSAGE_TYPE_TEXT {
		content = string(m.content)
	}
	output := fmt.Sprintf("%d%s|%s", m.msgType, m.author.name, content)
	return []byte(output)
}

func parseMsgType(firstByte byte) (int, error) {
	num, err := strconv.Atoi(string(firstByte))
	if err != nil {
		return 0, err
	}
	if num < 1 || num > 6 {
		return 0, errors.New("Invalid message type")
	}

	return num, nil
}

func parseMessageData(data []byte) (int, []byte, error) {
	msgType, err := parseMsgType(data[0])

	if err != nil {
		return 0, nil, err
	}

	content := data[1:]

	return msgType, content, nil
}

func NewMessage(author *Client, room *Room, msgType int, content []byte) Message {
	return Message{author: author, room: room, msgType: msgType, content: content}
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
	log.Printf("Broadcasting message %s", message.encode())
	h.broadcastChan <- message
}

func (h *Hub) handleRegister(client *Client, room *Room) {
	room.clients = append(room.clients, client)
	client.listen()
	log.Println("client connected ")
	log.Println("clients in the room: ", len(room.clients))
}

func (h *Hub) handleUnregister(client *Client, room *Room) {
	for i, c := range room.clients {
		if c == client {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}

	leaveMsg := NewMessage(client, room, MESSAGE_TYPE_LEAVE, nil)
	go h.broadcast(leaveMsg)

	client.close()

	if room.ClientCount() == 0 {
		h.terminateRoomIfEmpty(room)
	}

	log.Println("client disconnected ", client.name)
	log.Println("clients in the room: ", len(room.clients))
}

func (h *Hub) handleBroadcast(message Message) {
	encoded := message.encode()
	log.Printf("Handling broadcasted message: %s", encoded)

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
		if err != nil || messageType == websocket.CloseMessage {
			h.unregister(client, room)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		msgType, content, err := parseMessageData(message)
		if err != nil {
			log.Println("Invalid message received")
			continue
		}

		msg := NewMessage(client, room, msgType, content)

		if msg.msgType == MESSAGE_TYPE_NAME {
			client.setName(string(msg.content))
		}

		if client.name == "" && msg.msgType != MESSAGE_TYPE_NAME {
			log.Println("Client name not set")
			continue
		}

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
