package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetrics(t *testing.T) {
	server := NewMetricsServer(":2112")

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
