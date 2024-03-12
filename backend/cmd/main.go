package main

import (
	"github.com/olzhasar/gochat/pkg/chat"
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

	hub := chat.NewHub()
	hub.Run()

	server := chat.NewServer(hub)
	server.Run(port)
}
