package main

import (
	"github.com/olzhasar/gochat/pkg/metrics"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	metricsServer := metrics.NewServer("2112")
	metricsServer.Run()

	hub := NewHub()
	hub.Run()

	server := NewServer(hub)
	server.Run(port)
}
