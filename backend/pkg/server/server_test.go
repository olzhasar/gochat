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
	"github.com/olzhasar/gochat/pkg/server"
)

func TestCreateRoom(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	server := server.NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	roomId, err := createRoom(ts)
	if err != nil {
		t.Fatal(err)
	}

	room := hub.GetRoom(roomId)
	if room == nil {
		t.Fatal("expected room to be created")
	}
}

func TestCreateRoomConcurrent(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	server := server.NewServer(hub)

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
	hub := chat.NewHub()
	hub.Run()

	server := server.NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	roomId, err := createRoom(ts)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	for range 3 {
		wg.Go(func() {
			createRoom(ts)
		})
		wg.Go(func() {
			http.Get(ts.URL + "/room/" + roomId)
		})
	}

	wg.Wait()
}

func TestConnectToRoom(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	room := hub.CreateRoom()

	server := server.NewServer(hub)

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

func TestConnectToUnexistingRoom(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	server := server.NewServer(hub)

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
	hub.Run()

	room := hub.CreateRoom()

	server := server.NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts, room.ID)
	conn2 := makeConnection(ts, room.ID)
	defer conn1.Close()
	defer conn2.Close()

	conn1.WriteMessage(websocket.TextMessage, []byte("2test"))

	message := []byte("1hello")
	if err := conn1.WriteMessage(websocket.TextMessage, message); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	checkReceivedMessage(t, conn2, "2test|")
	checkReceivedMessage(t, conn2, "1test|hello")
}

func TestLeaveMessage(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	room := hub.CreateRoom()

	server := server.NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts, room.ID)
	defer conn1.Close()

	conn2 := makeConnection(ts, room.ID)
	conn2.WriteMessage(websocket.TextMessage, []byte("2leaver"))
	conn2.Close()

	checkReceivedMessage(t, conn1, "2leaver|")
	checkReceivedMessage(t, conn1, "3leaver|")
}

func TestGetRoom(t *testing.T) {
	hub := chat.NewHub()
	hub.Run()

	room := hub.CreateRoom()

	server := server.NewServer(hub)

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
	hub := chat.NewHub()
	hub.Run()

	server := server.NewServer(hub)

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

func makeConnection(ts *httptest.Server, roomId string) *websocket.Conn {
	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/ws/" + roomId

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
		return "", errors.New(fmt.Sprintf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	roomId := string(body)
	if roomId == "" {
		return "", errors.New("want roomId, got empty string")
	}

	return roomId, nil
}

func checkReceivedMessage(t testing.TB, conn *websocket.Conn, expected string) {
	t.Helper()

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	_, received, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("expected message %s, got error %s", expected, err)
	}

	if string(received) != expected {
		t.Fatalf("expected message %s, got %s", expected, received)
	}
}
