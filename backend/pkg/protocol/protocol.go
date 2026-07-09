package protocol

import (
	"errors"
	"github.com/ugorji/go/codec"
)

type MessageType int

const (
	MessageTypeText MessageType = iota
	MessageTypeJoin
	MessageTypeLeave
	MessageTypeStartTyping
	MessageTypeStopTyping
)

func (t MessageType) IsValid() bool {
	return t >= MessageTypeText && t <= MessageTypeStopTyping
}

type Message struct {
	Type       MessageType
	RoomID     string
	ClientID   string
	ClientName string
	Content    string
}

func (s *Message) Encode() []byte {
	var payload []byte
	var mh codec.MsgpackHandle
	encoder := codec.NewEncoderBytes(&payload, &mh)
	err := encoder.Encode(s)
	if err != nil {
		panic(err)
	}

	return payload
}

func Decode(payload []byte) (Message, error) {
	var msg Message
	var mh codec.MsgpackHandle
	decoder := codec.NewDecoderBytes(payload, &mh)
	err := decoder.Decode(&msg)

	if err != nil {
		return msg, err
	}

	if !msg.Type.IsValid() {
		return msg, errors.New("Invalid msg")
	}

	return msg, err
}
