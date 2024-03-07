package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	store    Store
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func readMessages(conn *websocket.Conn, store Store) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		err = store.AddMessage(string(message))
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	go readMessages(conn, s.store)
}

func NewServer(store Store) *Server {
	s := &Server{store: store}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/ws", s.handleWS)

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return s
}

func main() {
	store, err := NewSQLStore("main.db")
	if err != nil {
		log.Fatal(err)
	}
	server := NewServer(store)

	log.Fatal(http.ListenAndServe(":8080", server))
}
