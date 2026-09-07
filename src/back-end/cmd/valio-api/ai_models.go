package main

import (
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/application/queries"
	aiconfig "github.com/valio-projects/valio.code/internal/configuration/ai"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	"github.com/valio-projects/valio.code/internal/infrastructure/ai"
	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"os"
)

// configuredModels adapts validated operator configuration to query use cases.
type configuredModels struct {
	document    aiconfig.Document
	defaultName string
}

func (m configuredModels) ResolveModel(name string, kind embeddings.Kind) (embeddings.Profile, embeddings.Embedder, error) {
	if name == "" {
		name = m.defaultName
	}
	if name == "" || !kind.IsValid() {
		return embeddings.Profile{}, nil, fault.ErrInvalid
	}
	profile, provider, err := m.document.Resolve(name)
	if err != nil {
		return profile, nil, fault.ErrUnavailable
	}
	profile.Kind = kind
	profile.Fingerprint = embeddings.FingerprintProfile(profile)
	return profile, provider, profile.Validate()
}

// configureAI wires persistence even without a model, so disabled capability is
// explicit. Profiles are loaded once; configuration changes require a restart.
func configureAI(service *queries.Service, db *surreal.Client) error {
	vectors, err := surreal.NewEmbeddingRepository(db)
	if err != nil {
		return err
	}
	service.Vectors = vectors
	if path := os.Getenv("VALIO_AI_CONFIG"); path != "" {
		document, err := aiconfig.Load(path)
		if err != nil {
			return err
		}
		service.Models = configuredModels{document: document, defaultName: os.Getenv("VALIO_AI_DEFAULT_PROFILE")}
	}
	if endpoint := os.Getenv("VALIO_RERANK_URL"); endpoint != "" {
		service.Reranker = ai.CrossEncoderReranker{Config: ai.Config{Endpoint: endpoint, Token: os.Getenv("VALIO_RERANK_TOKEN"), AllowInsecureHTTP: os.Getenv("VALIO_RERANK_ALLOW_INSECURE_HTTP") == "true"}}
	}
	return nil
}
