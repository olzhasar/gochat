package main

import (
	"github.com/olzhasar/gochat/pkg/metrics"
	"github.com/olzhasar/gochat/pkg/server"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	metricsServer := metrics.NewServer("2112")
	metricsServer.Run()

	server := server.NewServer(nil)
	server.Run(port)
}
