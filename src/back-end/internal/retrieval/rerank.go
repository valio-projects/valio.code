package retrieval

import (
	"context"
	"fmt"
	"sort"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

// Reranker optionally assigns query-specific scores to already bounded candidates.
// Implementations may be local or remote; this package does not provide a model.
type Reranker interface {
	// Rerank returns one score per supplied chunk ID without changing source scope.
	Rerank(context.Context, string, []domainretrieval.Chunk) ([]RerankScore, error)
}

// RerankScore is a provider result for one candidate chunk.
type RerankScore struct {
	// ChunkID identifies the scored candidate.
	ChunkID string
	// Score is comparable only within the reranker invocation.
	Score float64
}

// ApplyReranker applies one optional reranker to existing candidates and keeps
// their previous provenance. It rejects missing, duplicate, and unknown scores.
func ApplyReranker(ctx context.Context, reranker Reranker, query string, candidates []ScoredChunk) ([]ScoredChunk, error) {
	if reranker == nil || len(candidates) == 0 {
		return append([]ScoredChunk(nil), candidates...), nil
	}
	chunks := make([]domainretrieval.Chunk, len(candidates))
	known := map[string]bool{}
	for i, candidate := range candidates {
		chunks[i], known[candidate.Chunk.ID] = candidate.Chunk, true
	}
	scores, err := reranker.Rerank(ctx, query, chunks)
	if err != nil {
		return nil, err
	}
	byID := map[string]float64{}
	seen := map[string]bool{}
	for _, score := range scores {
		if !known[score.ChunkID] || score.ChunkID == "" || seen[score.ChunkID] {
			return nil, fmt.Errorf("invalid reranker score set")
		}
		seen[score.ChunkID] = true
		byID[score.ChunkID] = score.Score
	}
	if len(byID) != len(candidates) {
		return nil, fmt.Errorf("reranker must score every candidate")
	}
	result := append([]ScoredChunk(nil), candidates...)
	sort.Slice(result, func(i, j int) bool {
		left, right := byID[result[i].Chunk.ID], byID[result[j].Chunk.ID]
		if left != right {
			return left > right
		}
		return result[i].Chunk.ID < result[j].Chunk.ID
	})
	for i := range result {
		result[i].Score = byID[result[i].Chunk.ID]
		result[i].Provenance = append(result[i].Provenance, ChannelScore{Channel: ChannelReranker, Score: result[i].Score, Rank: i + 1})
	}
	return result, nil
}
