package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Client struct {
	name string
	conn *websocket.Conn
}

func (c *Client) write(message []byte) {
	c.conn.WriteMessage(websocket.TextMessage, message)
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn}
}

type Server struct {
	mux      *http.ServeMux
	upgrader websocket.Upgrader
	clients  []*Client
}

func (s *Server) register(client *Client) {
	s.clients = append(s.clients, client)
}

func (s *Server) unregister(client *Client) {
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade connection")
		log.Println(err)
		return
	}

	log.Println("client connected")

	client := NewClient(conn)
	s.register(client)

	go s.readMessages(client)
}

func (s *Server) readMessages(client *Client) {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Println("client disconnected")
				s.unregister(client)
				return
			}
			log.Println("Unexpected error")
			log.Println(err)
			continue
		}

		log.Println("message received: ", string(message))

		for _, c := range s.clients {
			if c != client {
				c.write(message)
			}
		}
	}
}

func NewServer() *Server {
	s := &Server{}

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return s
}

func main() {
	server := NewServer()
	log.Fatal(http.ListenAndServe(":8080", server))
}
