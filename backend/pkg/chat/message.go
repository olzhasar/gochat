package chat

import (
	"errors"
	"fmt"
	"strconv"
)

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

func (m Message) Encode() []byte {
	var content string
	if m.msgType == MESSAGE_TYPE_TEXT {
		content = string(m.content)
	}
	output := fmt.Sprintf("%d%s|%s", m.msgType, m.author.name, content)
	return []byte(output)
}
func NewMessage(author *Client, room *Room, msgType int, content []byte) Message {
	return Message{author: author, room: room, msgType: msgType, content: content}
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
