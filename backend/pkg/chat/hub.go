package chat

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/olzhasar/gochat/pkg/metrics"
)

const EMPTY_ROOM_TIMEOUT = 1 * time.Minute

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

func (h *Hub) Register(client *Client, room *Room) {
	h.registerChan <- NewInstruction(client, room)
}

func (h *Hub) Unregister(client *Client, room *Room) {
	h.unregisterChan <- NewInstruction(client, room)
}

func (h *Hub) Broadcast(message Message) {
	h.broadcastChan <- message
}

func (h *Hub) handleRegister(client *Client, room *Room) {
	room.clients = append(room.clients, client)
	metrics.ClientCount.Inc()
	client.listen()
}

func (h *Hub) handleUnregister(client *Client, room *Room) {
	for i, c := range room.clients {
		if c == client {
			room.clients = append(room.clients[:i], room.clients[i+1:]...)
			break
		}
	}

	if client.name != "" {
		leaveMsg := NewMessage(client, room, MESSAGE_TYPE_LEAVE, nil)
		go h.Broadcast(leaveMsg)
	}

	client.close()

	h.scheduleRoomTermination(room)

	metrics.ClientCount.Dec()
}

func (h *Hub) handleBroadcast(message Message) {
	encoded := message.Encode()

	for _, client := range message.room.clients {
		if client != message.author {
			client.write(encoded)
			metrics.MessagesBroadcastedCount.Inc()
		}
	}

	metrics.MessagesReceivedCount.Inc()
}

func (h *Hub) Run() {
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

func (h *Hub) ListenClient(client *Client, room *Room) {
	for {
		messageType, message, err := client.conn.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			h.Unregister(client, room)
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		msgType, content, err := parseMessageData(message)
		if err != nil {
			log.Println("Invalid message received. Disconnecting client.")
			h.Unregister(client, room)
			continue
		}

		msg := NewMessage(client, room, msgType, content)

		if msg.msgType == MESSAGE_TYPE_NAME {
			client.setName(string(msg.content))
		}

		if client.name == "" && msg.msgType != MESSAGE_TYPE_NAME {
			log.Println("Client name not set. Disconnecting client.")
			h.Unregister(client, room)
			continue
		}

		h.Broadcast(msg)
	}
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

	h.scheduleRoomTermination(room)

	metrics.RoomCount.Inc()

	return room
}

func (h *Hub) GetRoom(id string) *Room {
	return h.rooms[id]
}

func (h *Hub) RoomCount() int {
	return len(h.rooms)
}

func (h *Hub) scheduleRoomTermination(room *Room) {
	go func() {
		time.AfterFunc(EMPTY_ROOM_TIMEOUT, func() {
			if room.ClientCount() > 0 || h.rooms[room.ID] == nil {
				return
			}
			delete(h.rooms, room.ID)
			metrics.RoomCount.Dec()
		})
	}()
}

func generateID() string {
	return uuid.New().String()
}
