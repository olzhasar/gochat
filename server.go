package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	hub      *Hub
}

func handleWS(upgrader websocket.Upgrader, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection")
			log.Println(err)
			return
		}

		client := NewClient(conn)
		hub.register(client)
	}
}

func NewServer(hub *Hub) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleWS(upgrader, hub))

	return &Server{upgrader: upgrader, hub: hub, mux: mux}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
