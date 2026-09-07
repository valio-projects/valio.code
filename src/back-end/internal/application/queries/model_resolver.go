package queries

import "github.com/valio-projects/valio.code/internal/domain/embeddings"

// ModelResolver selects only operator-configured providers. Query inputs never
// supply network endpoints or credentials. Kind selects a separate vector space.
type ModelResolver interface {
	ResolveModel(name string, kind embeddings.Kind) (embeddings.Profile, embeddings.Embedder, error)
}
