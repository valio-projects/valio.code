package ai

import (
	"context"
	"encoding/json"
	"errors"
	"math"
)

// CrossEncoderReranker calls one explicitly configured endpoint; callers never supply URLs.
type CrossEncoderReranker struct{ Config Config }

func (r CrossEncoderReranker) Rerank(ctx context.Context, query string, documents []string) ([]float32, error) {
	if len(documents) == 0 || len(documents) > 100 || len(query) > maxRequestBytes {
		return nil, errors.New("invalid rerank request")
	}
	data, err := request(ctx, r.Config, "", map[string]any{"query": query, "documents": documents})
	if err != nil {
		return nil, err
	}
	var output struct {
		Scores []float32 `json:"scores"`
	}
	if json.Unmarshal(data, &output) != nil || len(output.Scores) != len(documents) {
		return nil, errors.New("invalid rerank response")
	}
	for _, score := range output.Scores {
		if math.IsNaN(float64(score)) || math.IsInf(float64(score), 0) {
			return nil, errors.New("non-finite rerank score")
		}
	}
	return output.Scores, nil
}
