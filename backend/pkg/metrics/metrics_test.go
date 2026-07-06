package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olzhasar/gochat/pkg/metrics"
)

func TestServer(t *testing.T) {
	server := metrics.NewServer(":2112")

	ts := httptest.NewServer(server)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
