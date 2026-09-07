package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeEndpointReportsOnlyAvailability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("probe used a mutating request")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("model list is intentionally ignored"))
	}))
	defer server.Close()
	result := probeEndpoint(context.Background(), &http.Client{Timeout: time.Second}, "test", server.URL)
	if result.Status != ProbeAvailable || result.Provider != "test" || result.Endpoint != server.URL {
		t.Fatalf("probe result = %+v", result)
	}
	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer unavailable.Close()
	result = probeEndpoint(context.Background(), &http.Client{Timeout: time.Second}, "test", unavailable.URL)
	if result.Status != ProbeAbsent {
		t.Fatalf("non-2xx probe result = %+v", result)
	}
}
