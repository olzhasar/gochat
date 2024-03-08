package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Client struct {
	conn *websocket.Conn
}

func (c *Client) write(message []byte) {
	c.conn.WriteMessage(websocket.TextMessage, message)
}

type Registry interface {
	all() []*Client
	add(client *Client)
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn}
}

type MemoryRegistry struct {
	clients []*Client
}

func (m *MemoryRegistry) add(client *Client) {
	m.clients = append(m.clients, client)
}

func (m *MemoryRegistry) all() []*Client {
	return m.clients
}

func NewMemoryRegistry() *MemoryRegistry {
	return &MemoryRegistry{}
}

type Server struct {
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	store    Store
	registry Registry
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func readMessages(conn *websocket.Conn, store Store, registry Registry) {
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

		for _, client := range registry.all() {
			go client.write(message)
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn)
	s.registry.add(client)

	go readMessages(conn, s.store, s.registry)
}

func NewServer(store Store, registry Registry) *Server {
	s := &Server{store: store, registry: registry}

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

	registry := NewMemoryRegistry()

	server := NewServer(store, registry)

	log.Fatal(http.ListenAndServe(":8080", server))
}
