package queries

import (
	"context"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain/retrieval"
	ranking "github.com/valio-projects/valio.code/internal/retrieval"
	"math"
	"sort"
	"strings"
)

// Retrieve ranks scoped immutable chunks; verified source-search semantics
// remain on Search. Model failures are explicit and never invent semantic hits.
func (s Service) Retrieve(ctx context.Context, q RetrievalQuery) (RetrievalResult, error) {
	result := RetrievalResult{Mode: q.Mode, Hits: []ranking.ScoredChunk{}, Diagnostics: []string{}}
	if q.Mode == "" {
		q.Mode = "lexical"
		result.Mode = q.Mode
	}
	if len(q.Query) > 8192 || q.Limit < 0 || q.Limit > 100 || q.Query == "" && q.Mode != "structural" {
		return result, fault.ErrInvalid
	}
	switch q.Mode {
	case "lexical", "symbol", "structural", "semantic", "hybrid":
	default:
		return result, fault.ErrInvalid
	}
	if q.Limit == 0 {
		q.Limit = 20
	}
	v, chunks, err := s.scopedChunks(ctx, q.Scope)
	if err != nil {
		return result, err
	}
	result.ViewID = v.ID
	if q.Mode == "semantic" || q.Mode == "hybrid" {
		result.Diagnostics = append(result.Diagnostics, "SEMANTIC_SEARCH_COVERS_INDEXED_REPRESENTATIONS_ONLY")
	}
	var ranked []ranking.ScoredChunk
	switch q.Mode {
	case "lexical":
		ranked, err = ranking.RankBM25(q.Query, chunks, ranking.BM25Options{})
	case "symbol":
		ranked = symbolRank(q.Query, chunks)
	case "structural":
		var target retrieval.Chunk
		for _, c := range chunks {
			if c.ID == q.TargetChunkID {
				target = c
				break
			}
		}
		if target.ID == "" {
			return result, fault.ErrNotFound
		}
		ranked = []ranking.ScoredChunk{}
		for _, candidate := range chunks {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			if candidate.ID == target.ID {
				continue
			}
			score := ranking.StructuralSimilarity(target, candidate)
			if score > 0 {
				ranked = append(ranked, ranking.ScoredChunk{Chunk: candidate, Score: score})
			}
		}
		rankChannel(ranked, "structural")
		result.Diagnostics = append(result.Diagnostics, "STRUCTURAL_MATCH_IS_NOT_SEMANTIC_EQUIVALENCE")
	case "semantic":
		ranked, err = s.semanticRank(ctx, v.ID, q, chunks)
	case "hybrid":
		lexical, e := ranking.RankBM25(q.Query, chunks, ranking.BM25Options{})
		if e != nil {
			return result, fault.ErrInvalid
		}
		semantic, e := s.semanticRank(ctx, v.ID, q, chunks)
		if e != nil {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			result.Diagnostics = append(result.Diagnostics, "SEMANTIC_CHANNEL_UNAVAILABLE")
			semantic = nil
		}
		ranked = ranking.FuseRRF(60, lexical[:min(100, len(lexical))], symbolRank(q.Query, chunks), semantic)
	}
	if err != nil {
		return result, err
	}
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	result.CandidateCount = len(ranked)
	if q.Rerank {
		if s.Reranker == nil {
			return result, fault.ErrUnavailable
		}
		ranked = ranked[:min(100, len(ranked))]
		docs := make([]string, len(ranked))
		for i, c := range ranked {
			docs[i] = c.Chunk.Header + "\n" + c.Chunk.Text
		}
		if len(docs) > 0 {
			scores, e := s.Reranker.Rerank(ctx, q.Query, docs)
			if e != nil {
				if ctx.Err() != nil {
					return result, ctx.Err()
				}
				return result, fault.ErrUnavailable
			}
			if len(scores) != len(ranked) {
				return result, fault.ErrUnavailable
			}
			for i, score := range scores {
				if math.IsNaN(float64(score)) || math.IsInf(float64(score), 0) {
					return result, fault.ErrUnavailable
				}
				ranked[i].Score = float64(score)
				ranked[i].Provenance = append(ranked[i].Provenance, ranking.ChannelScore{Channel: "reranker", Score: float64(score)})
			}
			sort.Slice(ranked, func(i, j int) bool {
				if ranked[i].Score == ranked[j].Score {
					return ranked[i].Chunk.ID < ranked[j].Chunk.ID
				}
				return ranked[i].Score > ranked[j].Score
			})
			for i := range ranked {
				ranked[i].Provenance[len(ranked[i].Provenance)-1].Rank = i + 1
			}
		}
	}
	result.Truncated = len(ranked) > q.Limit || result.CandidateCount > len(ranked)
	result.Hits = ranked[:min(q.Limit, len(ranked))]
	return result, nil
}

func symbolRank(query string, chunks []retrieval.Chunk) []ranking.ScoredChunk {
	out := []ranking.ScoredChunk{}
	needle := strings.ToLower(strings.TrimSpace(query))
	if needle == "" {
		return out
	}
	for _, c := range chunks {
		if c.Text == "" {
			continue
		}
		for _, r := range c.Representations {
			if r.Kind != retrieval.RepresentationSymbol {
				continue
			}
			score := 0.0
			if strings.EqualFold(strings.TrimSpace(r.Text), needle) {
				score = 2
			} else if strings.Contains(strings.ToLower(r.Text), needle) {
				score = 1
			}
			if score > 0 {
				out = append(out, ranking.ScoredChunk{Chunk: c, Score: score})
			}
			break
		}
	}
	rankChannel(out, "symbol")
	return out
}

func rankChannel(items []ranking.ScoredChunk, channel ranking.Channel) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].Chunk.ID < items[j].Chunk.ID
		}
		return items[i].Score > items[j].Score
	})
	for i := range items {
		items[i].Provenance = []ranking.ChannelScore{{Channel: channel, Score: items[i].Score, Rank: i + 1}}
	}
}
