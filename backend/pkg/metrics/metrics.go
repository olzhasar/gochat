package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	RoomCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "room_count",
		Help: "The number of active rooms",
	})
	ClientCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "client_count",
		Help: "The number of active clients",
	})
	MessagesReceivedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_received",
		Help: "The number of messages received from all connections",
	})
	MessagesBroadcastedCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "messages_broadcasted",
		Help: "The number of messages broadcasted to all connections",
	})
)

type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{port: port}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func (s *Server) Run() {
	go func() {
		log.Println("Starting metrics server on port", s.port)
		log.Fatal(http.ListenAndServe(":"+s.port, s))
	}()
}
