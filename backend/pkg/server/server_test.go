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

	msg1 := protocol.Message{
		Type:       protocol.MessageTypeJoin,
		RoomID:     room.ID,
		ClientID:   "123",
		ClientName: "test",
	}

	msg2 := protocol.Message{
		Type:       protocol.MessageTypeText,
		RoomID:     room.ID,
		ClientID:   "123",
		ClientName: "test",
		Content:    "asflkj",
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

	checkReceivedMessage(t, conn2, encoded1)
	checkReceivedMessage(t, conn2, encoded2)
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

	msg1 := protocol.Message{
		Type:       protocol.MessageTypeJoin,
		ClientID:   "123",
		ClientName: "Vincent Vega",
	}

	msg2 := protocol.Message{
		Type:     protocol.MessageTypeLeave,
		ClientID: "123",
		RoomID:   room.ID,
	}

	encoded1 := msg1.Encode()
	encoded2 := msg2.Encode()

	conn2.WriteMessage(websocket.TextMessage, encoded1)
	conn2.WriteMessage(websocket.TextMessage, encoded2)
	conn2.Close()

	checkReceivedMessage(t, conn1, encoded1)
	checkReceivedMessage(t, conn1, encoded2)
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

func checkReceivedMessage(t testing.TB, conn *websocket.Conn, want []byte) {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	_, got, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected message %s, got error %s", want, err)
	}

	if string(got) != string(want) {
		t.Fatalf("expected message %s, got %s", want, got)
	}
}
