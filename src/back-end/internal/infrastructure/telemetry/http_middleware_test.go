package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPTraceExcludesQueriesAndCredentials(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer provider.Shutdown(context.Background())
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)
	handler := HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "secret-source" || r.Header.Get("Authorization") != "Bearer secret-token" {
			t.Error("application request changed")
		}
		w.WriteHeader(200)
	}))
	req := httptest.NewRequest("GET", "http://localhost/api/v1/types?query=secret-source", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("User-Agent", "secret-agent")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	spans := recorder.Ended()
	if len(spans) != 1 {
		t.Fatalf("spans=%d", len(spans))
	}
	attrs := fmt.Sprint(spans[0].Attributes())
	if strings.Contains(attrs, "secret") {
		t.Fatalf("sensitive attributes: %s", attrs)
	}
}
