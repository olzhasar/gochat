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

func handleRoomCreate(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := hub.CreateRoom()

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(room.ID))
	}
}

func handleWS(upgrader websocket.Upgrader, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := r.PathValue("room")
		if roomID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		room := hub.GetRoom(roomID)
		if room == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Failed to upgrade connection")
			log.Println(err)
			return
		}

		client := NewClient(conn)
		hub.register(client, room)

		hub.listenClient(client, room)
	}
}

func NewServer(hub *Hub) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /room", handleRoomCreate(hub))
	mux.HandleFunc("GET /room/{room}", handleWS(upgrader, hub))

	return &Server{upgrader: upgrader, hub: hub, mux: mux}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
