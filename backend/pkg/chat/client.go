package chat

import (
	"errors"

	"github.com/olzhasar/gochat/pkg/protocol"
)

type Client struct {
	id     string
	name   string
	room   *Room
	WriteQ chan protocol.Message
}

func newClient(ID string) Client {
	return Client{
		id:     ID,
		WriteQ: make(chan protocol.Message),
	}
}

func (c *Client) JoinRoom(room *Room) {
	if c.room != nil {
		panic("client is already in a room")
	}

	room.join(c)
	c.room = room
}

// enqueue a message for sending to this client. Does not block
func (c *Client) enqueue(payload protocol.Message) {
	c.WriteQ <- payload
}

func (c *Client) setName(name string) {
	c.name = name
}

func (c *Client) receive(room *Room, payload protocol.Message) error {
	switch payload.Type {
	case protocol.MessageTypeJoin:
		if c.name != "" {
			return errors.New("Name already set")
		}

		if payload.ClientName == "" {
			return errors.New("Required field Name is missing")
		}
		c.setName(payload.ClientName)

	default:
		if c.name == "" {
			return errors.New("Name has not been set")
		}
	}

	room.broadcast(payload)
	return nil
}

func (c *Client) close() {
	close(c.WriteQ)
}
