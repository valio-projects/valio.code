package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthChecksRunningAPIWithoutTokenOrDatabase(t *testing.T) {
	t.Setenv("VALIO_API_TOKEN", "")
	t.Setenv("VALIO_DB_PASSWORD", "")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" || r.Method != http.MethodGet {
			t.Errorf("unexpected readiness request %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("VALIO_API_ADDRESS", strings.TrimPrefix(server.URL, "http://"))
	if err := apiHealth(); err != nil {
		t.Fatal(err)
	}
}

func TestHealthRejectsUnreadyAndRedirects(t *testing.T) {
	for _, status := range []int{http.StatusServiceUnavailable, http.StatusFound} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Location", "https://example.invalid/")
				w.WriteHeader(status)
			}))
			defer server.Close()
			t.Setenv("VALIO_API_ADDRESS", strings.TrimPrefix(server.URL, "http://"))
			if err := apiHealth(); err == nil {
				t.Fatal("unready API was accepted")
			}
		})
	}
}
