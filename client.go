package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Name string
	conn *websocket.Conn
}

func (c *Client) write(message []byte) {
	c.conn.WriteMessage(websocket.TextMessage, message)
}

func (c *Client) setName(name string) {
	c.Name = name
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn}
}
