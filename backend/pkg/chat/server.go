package chat

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Server struct {
	mux             *http.ServeMux
	upgrader        websocket.Upgrader
	hub             *Hub
	corsAllowOrigin string
}

func (s *Server) handleRoomCreate(w http.ResponseWriter, r *http.Request) {
	room := s.hub.CreateRoom()

	s.setCORSPolicy(w)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(room.ID))
}

func (s *Server) handleRoomGet(w http.ResponseWriter, r *http.Request) {
	s.setCORSPolicy(w)

	roomID := r.PathValue("room")
	if roomID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room := s.hub.GetRoom(roomID)
	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("room")
	if roomID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	room := s.hub.GetRoom(roomID)
	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection")
		log.Println(err)
		return
	}

	client := NewClient(conn)
	s.hub.Register(client, room)

	s.hub.ListenClient(client, room)
}

func (s *Server) setCORSPolicy(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", s.corsAllowOrigin)
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	s.setCORSPolicy(w)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) configureRoutes() {
	s.mux.HandleFunc("OPTIONS /*", s.handleOptions)
	s.mux.HandleFunc("POST /room", s.handleRoomCreate)
	s.mux.HandleFunc("GET /room/{room}", s.handleRoomGet)
	s.mux.HandleFunc("GET /ws/{room}", s.handleWS)
}

func NewServer(hub *Hub) *Server {
	var corsAllowOrigin string
	if corsAllowOrigin = os.Getenv("CORS_ORIGIN"); corsAllowOrigin == "" {
		corsAllowOrigin = "*"
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if corsAllowOrigin == "*" {
				return true
			}
			origin := r.Header.Get("Origin")
			return origin == corsAllowOrigin
		},
	}
	mux := http.NewServeMux()

	server := &Server{upgrader: upgrader, hub: hub, corsAllowOrigin: corsAllowOrigin, mux: mux}
	server.configureRoutes()

	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) Run(port string) {
	log.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, s))
}
