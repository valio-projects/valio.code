package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

func profile() embeddings.Profile {
	p := embeddings.Profile{ID: "p", Kind: embeddings.Code, SchemaVersion: 1, Provider: "ollama", Model: "m", Revision: "r", Distance: "cosine", Dimension: 2}
	p.Fingerprint = embeddings.FingerprintProfile(p)
	return p
}
func TestOllamaRejectsMalformedDimensionAndStatus(t *testing.T) {
	for _, response := range []struct {
		status int
		body   string
	}{{200, `{"embeddings":[[1]]}`}, {500, `x`}} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(response.status)
			_, _ = w.Write([]byte(response.body))
		}))
		_, err := (OllamaEmbedder{Config: Config{Endpoint: server.URL, AllowInsecureHTTP: true}}).Embed(context.Background(), profile(), "safe")
		server.Close()
		if err == nil {
			t.Fatal("accepted invalid provider response")
		}
	}
}

func TestOllamaRequestsConfiguredDimensionWithoutTruncation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/embed" {
			t.Fatalf("unexpected Ollama request %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Model      string `json:"model"`
			Input      string `json:"input"`
			Dimensions int    `json:"dimensions"`
			Truncate   bool   `json:"truncate"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "m" || body.Input != "safe" || body.Dimensions != 2 || body.Truncate {
			t.Fatalf("unexpected embed body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"embeddings":[[1,2]]}`))
	}))
	defer server.Close()
	vector, err := (OllamaEmbedder{Config: Config{Endpoint: server.URL, AllowInsecureHTTP: true}}).Embed(context.Background(), profile(), "safe")
	if err != nil || len(vector) != 2 {
		t.Fatalf("embed result = %v, %v", vector, err)
	}
}
func TestOpenAIHonorsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (OpenAICompatibleEmbedder{Config: Config{Endpoint: server.URL, Timeout: time.Second, AllowInsecureHTTP: true, Token: "x"}}).Embed(ctx, profile(), "safe")
	if err == nil {
		t.Fatal("expected cancellation")
	}
}

func TestOpenAICompatibleAllowsLocalEndpointWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" || r.Header.Get("Authorization") != "" {
			t.Fatalf("unexpected local request: path=%q authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		var body struct {
			Model          string `json:"model"`
			Input          string `json:"input"`
			EncodingFormat string `json:"encoding_format"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.Model != "m" || body.Input != "local" || body.EncodingFormat != "float" {
			t.Fatalf("unexpected local embedding body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"data":[{"embedding":[1,2,3],"index":0}]}`))
	}))
	defer server.Close()
	local := profile()
	local.Dimension = 3
	vector, err := (OpenAICompatibleEmbedder{Config: Config{Endpoint: server.URL, AllowInsecureHTTP: true}}).Embed(context.Background(), local, "local")
	if err != nil || len(vector) != 3 {
		t.Fatalf("local embedding = %v, %v", vector, err)
	}
}

func TestEmbeddersAttachOptionalBearerToken(t *testing.T) {
	for _, provider := range []struct {
		name  string
		embed func(context.Context, embeddings.Profile) error
	}{
		{name: "ollama", embed: func(ctx context.Context, p embeddings.Profile) error {
			_, err := (OllamaEmbedder{Config: Config{Endpoint: tokenServerURL(t, `{"embeddings":[[1,2]]}`), AllowInsecureHTTP: true, Token: "local-token"}}).Embed(ctx, p, "safe")
			return err
		}},
		{name: "openai", embed: func(ctx context.Context, p embeddings.Profile) error {
			_, err := (OpenAICompatibleEmbedder{Config: Config{Endpoint: tokenServerURL(t, `{"data":[{"embedding":[1,2],"index":0}]}`), AllowInsecureHTTP: true, Token: "local-token"}}).Embed(ctx, p, "safe")
			return err
		}},
	} {
		t.Run(provider.name, func(t *testing.T) {
			if err := provider.embed(context.Background(), profile()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRequestRejectsMalformedTokenBeforeNetwork(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer server.Close()
	_, err := (OpenAICompatibleEmbedder{Config: Config{Endpoint: server.URL, AllowInsecureHTTP: true, Token: "bad token"}}).Embed(context.Background(), profile(), "safe")
	if err == nil || called {
		t.Fatalf("malformed token err=%v called=%t", err, called)
	}
}

func TestRequestRejectsRedirectBeforeCredentialRepost(t *testing.T) {
	targetCalls := 0
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { targetCalls++ }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-token" {
			t.Fatalf("source authorization = %q", r.Header.Get("Authorization"))
		}
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer source.Close()
	_, err := (OpenAICompatibleEmbedder{Config: Config{Endpoint: source.URL, AllowInsecureHTTP: true, Token: "local-token"}}).Embed(context.Background(), profile(), "safe")
	if !errors.Is(err, errRedirect) || targetCalls != 0 {
		t.Fatalf("redirect err=%v target calls=%d", err, targetCalls)
	}
}

func tokenServerURL(t *testing.T, response string) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer local-token" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return server.URL
}
