package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message struct {
	author  *Client
	content []byte
}

type Server struct {
	mux        *http.ServeMux
	upgrader   websocket.Upgrader
	clients    []*Client
	message    chan Message
	register   chan *Client
	unregister chan *Client
}

func NewServer() *Server {
	s := &Server{}

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	s.register = make(chan *Client)
	s.unregister = make(chan *Client)
	s.message = make(chan Message)

	return s
}

func composeMessage(author *Client, content []byte) string {
	return author.Name + "|" + string(content)
}

func (s *Server) registerClient(client *Client) {
	s.clients = append(s.clients, client)
}

func (s *Server) unregisterClient(client *Client) {
	for i, c := range s.clients {
		if c == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}
}

func (s *Server) handleMessage(message Message) {
	if message.author.Name == "" {
		message.author.setName(string(message.content))
		return
	}

	msg := composeMessage(message.author, message.content)

	for _, client := range s.clients {
		if client != message.author {
			client.write([]byte(msg))
		}
	}
}

func (s *Server) listen() {
	for {
		select {
		case message := <-s.message:
			s.handleMessage(message)
		case client := <-s.register:
			s.registerClient(client)
		case client := <-s.unregister:
			s.unregisterClient(client)
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
	s.register <- client

	go s.readMessages(client)
}

func (s *Server) readMessages(client *Client) {
	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Println("client disconnected")
				s.unregister <- client
				return
			}
			log.Println("Unexpected error")
			log.Println(err)
			continue
		}

		log.Println("message received: ", string(message))
		s.message <- Message{author: client, content: message}
	}
}
