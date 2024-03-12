package main

import (
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	metricsServer := NewMetricsServer("2112")
	metricsServer.Run()

	hub := NewHub()
	hub.Run()

	server := NewServer(hub)
	server.Run(port)
}
