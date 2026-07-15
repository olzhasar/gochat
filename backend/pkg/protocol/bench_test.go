package protocol_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/olzhasar/gochat/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

func BenchmarkEncode(b *testing.B) {
	msg := makeMessage(b)

	b.ResetTimer()

	for b.Loop() {
		msg.Encode()
	}
}

func BenchmarkDecode(b *testing.B) {
	msg := makeMessage(b)
	encoded := msg.Encode()

	got, err := protocol.Decode(encoded)
	if err != nil {
		b.Fatal(err)
	}

	if !proto.Equal(got, msg) {
		b.Fatal("decoded message does not match the original")
	}

	b.ResetTimer()

	for b.Loop() {
		protocol.Decode(encoded)
	}
}

func makeMessage(tb testing.TB) *protocol.Message {
	RoomID, err := uuid.NewUUID()
	if err != nil {
		tb.Fatal(err)
	}

	ClientID, err := uuid.NewUUID()
	if err != nil {
		tb.Fatal(err)
	}

	return &protocol.Message{
		Type:       protocol.MessageType_MSG_TEXT,
		RoomID:     RoomID.String(),
		AuthorID:   ClientID.String(),
		AuthorName: "foo",
		Content:    strings.Repeat("qwerty123", 32),
	}
}
