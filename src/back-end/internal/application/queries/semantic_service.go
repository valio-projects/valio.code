package queries

import (
	"context"
	appembeddings "github.com/valio-projects/valio.code/internal/application/embeddings"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	"github.com/valio-projects/valio.code/internal/domain/retrieval"
	ranking "github.com/valio-projects/valio.code/internal/retrieval"
)

func (s Service) semanticRank(ctx context.Context, viewID string, q RetrievalQuery, chunks []retrieval.Chunk) ([]ranking.ScoredChunk, error) {
	if s.Models == nil || s.Vectors == nil {
		return nil, fault.ErrUnavailable
	}
	kind := embeddings.Kind(q.Representation)
	if kind == "" {
		kind = embeddings.Code
	}
	if !kind.IsValid() {
		return nil, fault.ErrInvalid
	}
	profile, provider, err := s.Models.ResolveModel(q.ModelProfile, kind)
	if err != nil || provider == nil {
		return nil, fault.ErrUnavailable
	}
	records, err := s.Vectors.List(ctx, s.WorkspaceID, domain.ViewID(viewID), profile.Fingerprint, 1000)
	if err != nil {
		return nil, err
	}
	byID := map[string]retrieval.Chunk{}
	for _, c := range chunks {
		byID[c.ID] = c
	}
	scoped := []embeddings.Record{}
	for _, r := range records {
		if _, ok := byID[r.InputID]; ok {
			scoped = append(scoped, r)
		}
	}
	if len(scoped) == 0 {
		return nil, fault.ErrUnavailable
	}
	vector, err := provider.Embed(ctx, profile, q.Query)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fault.ErrUnavailable
	}
	scores, err := appembeddings.RankCosine(profile, vector, scoped, 100)
	if err != nil {
		return nil, fault.ErrUnavailable
	}
	hits := make([]ranking.ScoredChunk, 0, len(scores))
	for i, score := range scores {
		hits = append(hits, ranking.ScoredChunk{Chunk: byID[score.Record.InputID], Score: float64(score.Score), Provenance: []ranking.ChannelScore{{Channel: "semantic", Score: float64(score.Score), Rank: i + 1}}})
	}
	return hits, nil
}
