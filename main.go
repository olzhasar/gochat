package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	hub := NewHub()
	hub.run()

	server := NewServer(hub)

	log.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
