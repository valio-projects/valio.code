package telemetry

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestJaegerExport exercises the actual Compose OTLP receiver and query service.
func TestJaegerExport(t *testing.T) {
	endpoint := os.Getenv("VALIO_TEST_OTLP_URL")
	if endpoint == "" {
		t.Skip("requires Jaeger Compose test override")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	provider, err := NewProvider(ctx, Config{ServiceName: "valio-integration", Endpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	defer provider.Shutdown(context.Background())
	_, span := provider.TracerProvider.Tracer("compatibility").Start(ctx, "compose-export")
	traceID := span.SpanContext().TraceID().String()
	span.End()
	if err := provider.TracerProvider.ForceFlush(ctx); err != nil {
		t.Fatal(err)
	}
	client := http.Client{Timeout: time.Second}
	for ctx.Err() == nil {
		res, err := client.Get("http://127.0.0.1:16686/api/traces/" + traceID)
		if err == nil {
			var body struct {
				Data []json.RawMessage `json:"data"`
			}
			_ = json.NewDecoder(res.Body).Decode(&body)
			res.Body.Close()
			if res.StatusCode == http.StatusOK && len(body.Data) > 0 {
				return
			}
		}
		select {
		case <-ctx.Done():
		case <-time.After(100 * time.Millisecond):
		}
	}
	t.Fatal("exported trace was not returned by Jaeger query API")
}
