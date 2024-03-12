package chat

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	name          string
	conn          *websocket.Conn
	broadcastChan chan []byte
}

func (c *Client) write(message []byte) {
	c.broadcastChan <- message
}

func (c *Client) setName(name string) {
	c.name = name
}

func (c *Client) listen() {
	go func() {
		for msg := range c.broadcastChan {
			c.conn.WriteMessage(websocket.TextMessage, msg)
		}
	}()
}

func (c *Client) close() {
	close(c.broadcastChan)
	c.conn.Close()
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, broadcastChan: make(chan []byte)}
}
