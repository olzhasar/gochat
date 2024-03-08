package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebsocketConnection(t *testing.T) {
	server := NewServer()
	go server.listen()

	ts := httptest.NewServer(server)
	defer ts.Close()

	fmt.Println(ts.URL)

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial("ws"+ts.URL[4:], nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status code %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}

	time.Sleep(50 * time.Millisecond)

	if len(server.clients) != 1 {
		t.Fatalf("expected 1 client, got %d", len(server.clients))
	}
}

func TestSetName(t *testing.T) {
	server := NewServer()
	go server.listen()

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn := makeConnection(ts)
	defer conn.Close()

	name := "test"
	if err := conn.WriteMessage(websocket.TextMessage, []byte(name)); err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)

	if server.clients[0].name != name {
		t.Fatalf("expected name %s, got %s", name, server.clients[0].name)
	}
}

func TestWebsocketMessage(t *testing.T) {
	server := NewServer()
	go server.listen()

	ts := httptest.NewServer(server)
	defer ts.Close()

	conn1 := makeConnection(ts)
	conn2 := makeConnection(ts)
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

func makeConnection(ts *httptest.Server) *websocket.Conn {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws"+ts.URL[4:], nil)

	if err != nil {
		panic(err)
	}

	return conn
}
