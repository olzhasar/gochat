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

	server := NewServer()
	go server.listen()

	log.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
