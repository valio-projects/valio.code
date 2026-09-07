package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

// OllamaEmbedder calls Ollama's documented /api/embed endpoint. It requests the
// profile's declared output width and disables silent input truncation.
type OllamaEmbedder struct{ Config Config }

func (o OllamaEmbedder) Embed(ctx context.Context, profile embeddings.Profile, input string) ([]float32, error) {
	data, err := request(ctx, o.Config, "/api/embed", map[string]any{"model": profile.Model, "input": input, "dimensions": profile.Dimension, "truncate": false})
	if err != nil {
		return nil, err
	}
	var output struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if json.Unmarshal(data, &output) != nil || len(output.Embeddings) != 1 {
		return nil, errors.New("invalid Ollama embedding response")
	}
	return vector(output.Embeddings[0], profile.Dimension)
}
