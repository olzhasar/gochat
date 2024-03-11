package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	roomCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_count",
		Help: "The number of active rooms",
	})
	clientCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "client_count",
		Help: "The number of active clients",
	})
	messagesReceivedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_received",
		Help: "The number of messages received from all connections",
	})
	messagesBroadcastedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_broadcasted",
		Help: "The number of messages broadcasted to all connections",
	})
)

type MetricsServer struct {
	port string
}

func NewMetricsServer(port string) *MetricsServer {
	return &MetricsServer{port: port}
}

func (s *MetricsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func (s *MetricsServer) run() {
	log.Println("Starting metrics server on port", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, s))
}
