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

	metricsServer := NewMetricsServer("2112")
	go metricsServer.run()

	log.Println("Starting server on port", port)
	log.Fatal(http.ListenAndServe(":"+port, server))
}
