package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestCreateRoom(t *testing.T) {
	hub := NewHub()
	hub.run()

	server := NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	url := ts.URL + "/room"

	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	roomId := string(body)

	if roomId == "" {
		t.Fatal("expected non-empty room id")
	}

	room := hub.GetRoom(roomId)
	if room == nil {
		t.Fatal("expected room to be created")
	}
}

func TestConnectToRoom(t *testing.T) {
	hub := NewHub()
	hub.run()

	room := hub.CreateRoom()

	server := NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/room/" + room.ID

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
	hub := NewHub()
	hub.run()

	server := NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/room/123"

	_, resp, err := dialer.Dial(url, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestSetName(t *testing.T) {
	hub := NewHub()
	hub.run()

	room := hub.CreateRoom()

	server := NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn := makeConnection(ts, room.ID)
	defer conn.Close()

	name := "test"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(name)); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	if room.clients[0].name != name {
		t.Fatalf("expected name %s, got %s", name, room.clients[0].name)
	}
}

func TestRoomMessage(t *testing.T) {
	hub := NewHub()
	hub.run()

	room := hub.CreateRoom()

	server := NewServer(hub)

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts, room.ID)
	conn2 := makeConnection(ts, room.ID)
	defer conn1.Close()
	defer conn2.Close()

	conn1.WriteMessage(websocket.TextMessage, []byte("test"))

	message := []byte("hello")
	if err := conn1.WriteMessage(websocket.TextMessage, message); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	expected := "test|hello"

	_, received, err := conn2.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}

	if string(received) != expected {
		t.Fatalf("expected message %s, got %s", expected, received)
	}
}

func makeConnection(ts *httptest.Server, roomId string) *websocket.Conn {
	dialer := websocket.Dialer{}
	url := "ws" + ts.URL[4:] + "/room/" + roomId

	conn, _, err := dialer.Dial(url, nil)

	if err != nil {
		panic(err)
	}

	return conn
}