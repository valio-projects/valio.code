package surreal

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestBoundParametersAndScope(t *testing.T) {
	dangerous := "'; DELETE project; -- secret"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Surreal-DB") != "workspace_test" {
			t.Error("scope not bound")
		}
		if r.Header.Get("Content-Type") != "application/cbor" || r.Header.Get("Accept") != "application/cbor" {
			t.Error("RPC must use CBOR in both directions")
		}
		var body struct {
			Params []any `cbor:"params"`
		}
		data, _ := io.ReadAll(r.Body)
		if err := cbor.Unmarshal(data, &body); err != nil {
			t.Error(err)
			return
		}
		if len(body.Params) != 2 {
			t.Error("missing bound parameters")
			return
		}
		if strings.Contains(body.Params[0].(string), dangerous) {
			t.Error("data interpolated in SQL")
		}
		writeRPCFixture(t, w, "OK", []any{})
	}))
	defer srv.Close()
	c, err := New(Config{Endpoint: srv.URL, Namespace: "valio", Database: "workspace_test"})
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Put(context.Background(), "project", "one", map[string]string{"name": dangerous}); err != nil {
		t.Fatal(err)
	}
}
func TestDatabaseErrorDoesNotExposePayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeRPCFixture(t, w, "ERR", "secret source content")
	}))
	defer srv.Close()
	c, _ := New(Config{Endpoint: srv.URL, Namespace: "valio", Database: "test"})
	_, err := c.Query(context.Background(), "RETURN 1;", nil)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func writeRPCFixture(t *testing.T, w http.ResponseWriter, status string, result any) {
	t.Helper()
	data, err := cbor.Marshal(map[string]any{"id": "query", "result": []any{map[string]any{"status": status, "result": result}}})
	if err != nil {
		t.Error(err)
		return
	}
	w.Header().Set("Content-Type", "application/cbor")
	_, _ = w.Write(data)
}
