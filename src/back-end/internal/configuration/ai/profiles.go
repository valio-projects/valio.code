// Package ai loads explicitly selected external embedding profiles without network I/O.
package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	infra "github.com/valio-projects/valio.code/internal/infrastructure/ai"
	"github.com/valio-projects/valio.code/internal/validation"
)

// Document is the versioned, strict JSON profile file.
type Document struct {
	SchemaVersion int       `json:"schemaVersion"`
	Profiles      []Profile `json:"profiles"`
}

// Profile configures one model space; it never includes a token value.
type Profile struct {
	ID                string          `json:"id"`
	ProviderKind      string          `json:"providerKind"`
	URL               string          `json:"url"`
	Model             string          `json:"model"`
	Revision          string          `json:"revision"`
	Quantization      string          `json:"quantization"`
	Dimension         int             `json:"dimension"`
	Timeout           string          `json:"timeout"`
	TokenEnv          string          `json:"tokenEnv,omitempty"`
	AllowInsecureHTTP bool            `json:"allowInsecureHTTP,omitempty"`
	Kind              embeddings.Kind `json:"kind"`
	SchemaVersion     int             `json:"schemaVersion"`
}

// Load strictly decodes one profile document, rejecting unknown fields and trailing JSON.
func Load(path string) (Document, error) {
	info, e := os.Stat(path)
	if e != nil || info.Size() > 1<<20 {
		return Document{}, errors.New("invalid AI profile file")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		return Document{}, e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var out Document
	if e = d.Decode(&out); e != nil {
		return Document{}, errors.New("invalid AI profile JSON")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return Document{}, errors.New("invalid AI profile trailing JSON")
	}
	if out.SchemaVersion != 1 || len(out.Profiles) == 0 || len(out.Profiles) > 64 {
		return Document{}, errors.New("invalid AI profile document")
	}
	seen := map[string]bool{}
	for _, p := range out.Profiles {
		if seen[p.ID] || p.Validate() != nil {
			return Document{}, errors.New("invalid AI profile")
		}
		seen[p.ID] = true
	}
	return out, nil
}

// Resolve returns a validated domain profile and adapter without contacting the provider.
func (d Document) Resolve(name string) (embeddings.Profile, embeddings.Embedder, error) {
	for _, p := range d.Profiles {
		if p.ID == name {
			return p.resolve()
		}
	}
	return embeddings.Profile{}, nil, errors.New("AI profile not found")
}
func (p Profile) Validate() error {
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,79}$`).MatchString(p.ID) || !p.Kind.IsValid() || p.SchemaVersion < 1 || p.Dimension < 1 || p.Dimension > 16384 || p.Model == "" || p.Revision == "" || p.Quantization == "" {
		return errors.New("invalid AI profile")
	}
	if p.ProviderKind != "ollama" && p.ProviderKind != "openai-compatible" {
		return errors.New("invalid AI provider")
	}
	if _, e := validation.Endpoint(p.URL, !p.AllowInsecureHTTP); e != nil {
		return e
	}
	if p.TokenEnv != "" && !validEnv(p.TokenEnv) {
		return errors.New("invalid AI token environment")
	}
	if duration, e := time.ParseDuration(p.Timeout); e != nil || duration <= 0 || duration > 120*time.Second {
		return errors.New("invalid AI timeout")
	}
	return nil
}
func (p Profile) resolve() (embeddings.Profile, embeddings.Embedder, error) {
	if e := p.Validate(); e != nil {
		return embeddings.Profile{}, nil, e
	}
	timeout, _ := time.ParseDuration(p.Timeout)
	token := ""
	if p.TokenEnv != "" {
		token = os.Getenv(p.TokenEnv)
		if e := validation.Token(token, 1); e != nil {
			return embeddings.Profile{}, nil, e
		}
	}
	domain, e := embeddings.NewProfile(p.ID, p.Kind, p.SchemaVersion, p.ProviderKind, p.Model, p.Revision+";quant="+p.Quantization, "cosine", p.Dimension)
	if e != nil {
		return embeddings.Profile{}, nil, e
	}
	cfg := infra.Config{Endpoint: p.URL, Token: token, Timeout: timeout, AllowInsecureHTTP: p.AllowInsecureHTTP}
	if p.ProviderKind == "ollama" {
		return domain, infra.OllamaEmbedder{Config: cfg}, nil
	}
	return domain, infra.OpenAICompatibleEmbedder{Config: cfg}, nil
}
func validEnv(v string) bool {
	if len(v) == 0 || len(v) > 128 {
		return false
	}
	for i, c := range v {
		if !(c == '_' || (c >= 'A' && c <= 'Z') || (i > 0 && c >= '0' && c <= '9')) {
			return false
		}
	}
	return strings.TrimSpace(v) == v
}
