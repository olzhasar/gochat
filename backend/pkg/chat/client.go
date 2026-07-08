package chat

import (
	"log/slog"

	"github.com/gorilla/websocket"
)

type Client struct {
	name          string
	conn          *websocket.Conn
	room          *Room
	broadcastChan chan []byte
}

func NewClient(conn *websocket.Conn) *Client {
	if conn == nil {
		panic("no connection")
	}
	return &Client{conn: conn, broadcastChan: make(chan []byte)}
}

func (c *Client) JoinRoom(room *Room) {
	if c.room != nil {
		panic("client is already in a room")
	}

	room.join(c)
	c.room = room
}

// Send a message to this client. Does not block
func (c *Client) Send(message []byte) {
	c.broadcastChan <- message
}

func (c *Client) setName(name string) {
	c.name = name
}

func (c *Client) listenWS() {
	for {
		messageType, message, err := c.conn.ReadMessage()
		if err != nil || messageType == websocket.CloseMessage {
			if c.room != nil {
				c.room.leave(c)
				c.close()
			}
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		msgType, content, err := parseMessageData(message)
		if err != nil {
			slog.Error("Invalid message received. Disconnecting client.")
			if c.room != nil {
				c.room.leave(c)
			}
			continue
		}

		msg := NewMessage(c, c.room, msgType, content)

		if msg.msgType == MESSAGE_TYPE_NAME {
			c.setName(string(msg.content))
		}

		if c.name == "" && msg.msgType != MESSAGE_TYPE_NAME {
			slog.Info("Client name not set. Disconnecting client.")
			if c.room != nil {
				c.room.leave(c)
			}
			continue
		}

		msg.room.Broadcast(msg.Encode())
	}
}

func (c *Client) Run() {
	go func() {
		for msg := range c.broadcastChan {
			c.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()

	go c.listenWS()
}

func (c *Client) close() {
	close(c.broadcastChan)
	err := c.conn.Close()
	if err != nil {
		slog.Error("Failed to close connection")
	}
}
