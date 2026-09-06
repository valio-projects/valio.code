package surreal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBoundParametersAndScope(t *testing.T) {
	dangerous := "'; DELETE project; -- secret"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Surreal-DB") != "workspace_test" {
			t.Error("scope not bound")
		}
		var body struct {
			Params []json.RawMessage `json:"params"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if strings.Contains(string(body.Params[0]), dangerous) {
			t.Error("data interpolated in SQL")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":[{"status":"OK","result":[]}]}`))
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
		_, _ = w.Write([]byte(`{"result":[{"status":"ERR","result":"secret source content"}]}`))
	}))
	defer srv.Close()
	c, _ := New(Config{Endpoint: srv.URL, Namespace: "valio", Database: "test"})
	_, err := c.Query(context.Background(), "RETURN 1;", nil)
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("unsafe error: %v", err)
	}
}
