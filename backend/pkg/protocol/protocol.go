package protocol

import (
	"errors"

	"google.golang.org/protobuf/proto"
)

func (t MessageType) IsValid() bool {
	return t >= MessageType_MSG_TEXT && t <= MessageType_MSG_STOP_TYPING
}

func (s *Message) Encode() []byte {
	payload, err := proto.Marshal(s)
	if err != nil {
		panic(err)
	}

	return payload
}

func Decode(payload []byte) (*Message, error) {
	var msg Message
	err := proto.Unmarshal(payload, &msg)

	if err != nil {
		return nil, err
	}

	if !msg.Type.IsValid() {
		return nil, errors.New("Invalid msg")
	}

	return &msg, err
}
