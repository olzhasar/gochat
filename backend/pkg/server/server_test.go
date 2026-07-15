package server_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olzhasar/gochat/pkg/chat"
	"github.com/olzhasar/gochat/pkg/protocol"
	"github.com/olzhasar/gochat/pkg/server"
)

func TestCreateRoom(t *testing.T) {
	server := server.NewServer(nil)

	ts := httptest.NewServer(server)
	defer ts.Close()

	roomID, err := createRoom(ts)
	if err != nil {
		t.Fatal(err)
	}

	wantStatus := http.StatusNoContent
	resp, err := http.Get(ts.URL + "/room/" + roomID)
	if err != nil {
		t.Fatalf("want %d, got error: %s", wantStatus, err)
	}

	if resp.StatusCode != wantStatus {
		t.Fatalf("want %d, got %d", wantStatus, resp.StatusCode)
	}
}

func TestCreateRoomConcurrent(t *testing.T) {
	server := server.NewServer(nil)

	ts := httptest.NewServer(server)
	defer ts.Close()

	n_requests := 3
	err_ch := make(chan error, n_requests)

	for range n_requests {
		go func() {
			_, err := createRoom(ts)
			err_ch <- err
		}()
	}

	for range n_requests {
		if err := <-err_ch; err != nil {
			t.Fatalf("Error: %v\n", err)
		}
	}

}

func TestCreateAndGetRoomConcurrent(t *testing.T) {
	server := server.NewServer(nil)

	ts := httptest.NewServer(server)
	defer ts.Close()

	roomID, err := createRoom(ts)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for range 3 {
		wg.Go(func() {
			createRoom(ts)
		})
		wg.Go(func() {
			http.Get(ts.URL + "/room/" + roomID)
		})
	}

	wg.Wait()
}

func TestConnectToRoom(t *testing.T) {
	hub := chat.NewHub()
	server := server.NewServer(hub)

	room := hub.CreateRoom()

	ts := httptest.NewServer(server)
	defer ts.Close()

	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/ws/" + room.ID

	conn, resp, err := dialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status code %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(50 * time.Millisecond)

	if room.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", room.ClientCount())
	}
}

func TestConnectToNonExistentRoom(t *testing.T) {
	server := server.NewServer(nil)

	ts := httptest.NewServer(server)
	defer ts.Close()

	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/ws/123"

	_, resp, err := dialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestTextMessage(t *testing.T) {
	hub := chat.NewHub()
	server := server.NewServer(hub)

	room := hub.CreateRoom()

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts, room.ID)
	conn2 := makeConnection(ts, room.ID)
	defer conn1.Close()
	defer conn2.Close()

	name := "test"

	msg1 := protocol.Message{
		Type:    protocol.MessageType_MSG_JOIN,
		RoomID:  room.ID,
		Content: name,
	}

	msg2 := protocol.Message{
		Type:    protocol.MessageType_MSG_TEXT,
		RoomID:  room.ID,
		Content: "foo",
	}

	encoded1 := msg1.Encode()
	encoded2 := msg2.Encode()

	if err := conn1.WriteMessage(websocket.TextMessage, encoded1); err != nil {
		t.Fatal(err)
	}

	if err := conn1.WriteMessage(websocket.TextMessage, encoded2); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	got1 := receiveMessage(t, conn2)
	assertMessageType(t, got1, protocol.MessageType_MSG_JOIN)
	assertMessageAuthorName(t, got1, name)

	got2 := receiveMessage(t, conn2)
	assertMessageType(t, got2, protocol.MessageType_MSG_TEXT)
	assertMessageAuthorName(t, got2, name)
	assertMessageContent(t, got2, "foo")
}

func TestLeaveMessage(t *testing.T) {
	hub := chat.NewHub()
	server := server.NewServer(hub)

	room := hub.CreateRoom()

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts, room.ID)
	defer conn1.Close()

	conn2 := makeConnection(ts, room.ID)

	name := "Vincent Vega"

	msg1 := protocol.Message{
		Type:    protocol.MessageType_MSG_JOIN,
		Content: name,
	}

	msg2 := protocol.Message{
		Type:   protocol.MessageType_MSG_LEAVE,
		RoomID: room.ID,
	}

	encoded1 := msg1.Encode()
	encoded2 := msg2.Encode()

	conn2.WriteMessage(websocket.TextMessage, encoded1)
	conn2.WriteMessage(websocket.TextMessage, encoded2)
	conn2.Close()

	got1 := receiveMessage(t, conn1)
	assertMessageType(t, got1, protocol.MessageType_MSG_JOIN)
	assertMessageAuthorName(t, got1, name)

	got2 := receiveMessage(t, conn1)
	assertMessageType(t, got2, protocol.MessageType_MSG_LEAVE)
	assertMessageAuthorName(t, got2, name)
}

func TestGetRoom(t *testing.T) {
	hub := chat.NewHub()
	server := server.NewServer(hub)

	room := hub.CreateRoom()

	ts := httptest.NewServer(server)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/room/" + room.ID)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetUnexistingRoom(t *testing.T) {
	server := server.NewServer(nil)

	ts := httptest.NewServer(server)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/room/123")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func makeConnection(ts *httptest.Server, roomID string) *websocket.Conn {
	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/ws/" + roomID

	conn, _, err := dialer.Dial(url, nil)

	if err != nil {
		panic(err)
	}

	return conn
}

func createRoom(ts *httptest.Server) (string, error) {
	url := ts.URL + "/room"

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	roomID := string(body)
	if roomID == "" {
		return "", errors.New("want roomID, got empty string")
	}

	return roomID, nil
}

func receiveMessage(t testing.TB, conn *websocket.Conn) *protocol.Message {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	wsType, got, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	if wsType != websocket.TextMessage {
		t.Fatalf("unexpected ws message type: %d", wsType)
	}

	msg, err := protocol.Decode(got)
	if err != nil {
		t.Fatal(err)
	}

	return msg
}

func assertMessageAuthorName(t testing.TB, msg *protocol.Message, want string) {
	t.Helper()
	if msg.AuthorName != want {
		t.Fatalf("want name %s, got %s\n", want, msg.AuthorName)
	}
}

func assertMessageType(t testing.TB, msg *protocol.Message, want protocol.MessageType) {
	t.Helper()
	if msg.Type != want {
		t.Fatalf("want type %s, got %s\n", want, msg.Type)
	}
}

func assertMessageContent(t testing.TB, msg *protocol.Message, want string) {
	t.Helper()
	if msg.Content != want {
		t.Fatalf("want content %s, got %s\n", want, msg.Content)
	}
}
