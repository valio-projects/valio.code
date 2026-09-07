package ai

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valio-projects/valio.code/internal/domain/embeddings"
)

// OpenAICompatibleEmbedder calls a configured OpenAI-compatible /embeddings endpoint.
// A bearer token is optional because local servers such as LM Studio commonly do not use one.
type OpenAICompatibleEmbedder struct{ Config Config }

func (o OpenAICompatibleEmbedder) Embed(ctx context.Context, profile embeddings.Profile, input string) ([]float32, error) {
	data, err := request(ctx, o.Config, "/embeddings", map[string]any{"model": profile.Model, "input": input, "encoding_format": "float"})
	if err != nil {
		return nil, err
	}
	var output struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if json.Unmarshal(data, &output) != nil || len(output.Data) != 1 || output.Data[0].Index != 0 {
		return nil, errors.New("invalid OpenAI-compatible embedding response")
	}
	return vector(output.Data[0].Embedding, profile.Dimension)
}
