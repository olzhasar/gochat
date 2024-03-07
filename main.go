package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	upgrader websocket.Upgrader
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)

	mux.ServeHTTP(w, r)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()
}

func NewServer() *Server {
	s := &Server{}
	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return s
}

func main() {
	server := NewServer()

	log.Fatal(http.ListenAndServe(":8080", server))
}
