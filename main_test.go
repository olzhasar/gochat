package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWebsocketConnection(t *testing.T) {
	server := NewServer()
	ts := httptest.NewServer(server)
	defer ts.Close()

	fmt.Println(ts.URL)

	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial("ws"+ts.URL[4:]+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected status code %d, got %d", http.StatusSwitchingProtocols, resp.StatusCode)
	}
}

func TestWebsocketMessage(t *testing.T) {
	server := NewServer()
	ts := httptest.NewServer(server)
	defer ts.Close()

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws"+ts.URL[4:]+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	message := []byte("hello")
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		t.Fatal(err)
	}
}
