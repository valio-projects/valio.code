package embeddings

import "context"

// Embedder is an external model adapter. Implementations must bound transport and honor ctx.
type Embedder interface {
	// Embed returns a vector matching profile.Dimension for the policy-approved representation.
	Embed(ctx context.Context, profile Profile, representation string) ([]float32, error)
}

// Reranker is an optional external cross-encoder adapter; it does not create vectors.
type Reranker interface {
	// Rerank returns one score per candidate in the same order, or an error.
	Rerank(ctx context.Context, query string, candidates []string) ([]float32, error)
}
