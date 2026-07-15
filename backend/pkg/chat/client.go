package chat

import (
	"errors"

	"github.com/olzhasar/gochat/pkg/protocol"
)

type Client struct {
	id     string
	name   string
	room   *Room
	WriteQ chan *protocol.Message
}

func newClient(ID string) Client {
	return Client{
		id:     ID,
		WriteQ: make(chan *protocol.Message),
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
func (c *Client) enqueue(msg *protocol.Message) {
	c.WriteQ <- msg
}

func (c *Client) setName(name string) {
	c.name = name
}

func (c *Client) composeMessage(msgType protocol.MessageType, roomID string, content string) *protocol.Message {
	return &protocol.Message{
		Type:       msgType,
		AuthorID:   c.id,
		AuthorName: c.name,
		RoomID:     roomID,
		Content:    content,
	}
}

func (c *Client) addMsgIdentity(msg *protocol.Message) {
	msg.AuthorID = c.id
	msg.AuthorName = c.name
}

func (c *Client) receive(room *Room, msg *protocol.Message) error {
	switch msg.Type {
	case protocol.MessageType_MSG_JOIN:
		if c.name != "" {
			return errors.New("Name already set")
		}

		if msg.Content == "" {
			return errors.New("Required field Name is missing")
		}
		c.setName(msg.Content)

	default:
		if c.name == "" {
			return errors.New("Name has not been set")
		}
	}

	c.addMsgIdentity(msg)

	room.broadcast(msg)
	return nil
}

func (c *Client) close() {
	close(c.WriteQ)
}
