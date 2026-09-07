// Package embeddings computes vectors only from policy-approved supplied representations.
package embeddings

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
)

// Service coordinates idempotent vector computation; it neither reads source nor builds database queries.
type Service struct {
	// Repository finds and persists immutable records in their workspace/view scope.
	Repository repositories.EmbeddingRepository
	// Provider generates a vector from one policy-approved representation.
	Provider embeddings.Embedder
	// Now supplies record timestamps and defaults to time.Now when nil.
	Now func() time.Time
}

// Compute embeds one pre-approved representation and persists its immutable result.
func (s Service) Compute(ctx context.Context, profile embeddings.Profile, input embeddings.Input) (embeddings.Record, error) {
	if s.Repository == nil {
		return embeddings.Record{}, errors.New("embedding repository required")
	}
	if err := profile.Validate(); err != nil {
		return embeddings.Record{}, err
	}
	if err := input.Validate(); err != nil {
		return embeddings.Record{}, err
	}
	if profile.Kind != input.Kind {
		return embeddings.Record{}, errors.New("profile and input kind differ")
	}
	if existing, ok, err := s.Repository.Find(ctx, input.WorkspaceID, input.ViewID, input.ID, profile.Fingerprint, input.Fingerprint); err != nil {
		return embeddings.Record{}, err
	} else if ok && existing.Completeness != embeddings.MissingProvider {
		return existing, nil
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	if s.Provider == nil {
		r, err := embeddings.MissingRecord(profile, input, now())
		if err != nil {
			return embeddings.Record{}, err
		}
		return r, nil
	}
	vector, err := s.Provider.Embed(ctx, profile, input.Representation)
	if err != nil {
		return embeddings.Record{}, err
	}
	record, err := embeddings.NewRecord(profile, input, vector, now())
	if err != nil {
		return embeddings.Record{}, err
	}
	if err = s.Repository.Save(ctx, record); err != nil {
		// Concurrent callers may compute the same deterministic record. Read the
		// immutable winner rather than treating that normal race as a failed compute.
		cached, found, findErr := s.Repository.Find(ctx, input.WorkspaceID, input.ViewID, input.ID, profile.Fingerprint, input.Fingerprint)
		if findErr == nil && found && cached.Completeness != embeddings.MissingProvider {
			return cached, nil
		}
		return embeddings.Record{}, err
	}
	return record, nil
}

// RankCosine exactly scores a bounded candidate list. It deliberately has no ANN/HNSW claim.
func RankCosine(profile embeddings.Profile, query []float32, candidates []embeddings.Record, limit int) ([]Scored, error) {
	if err := profile.Validate(); err != nil {
		return nil, err
	}
	if limit < 1 || limit > 1000 {
		return nil, errors.New("invalid ranking limit")
	}
	if len(candidates) > 1000 {
		return nil, errors.New("candidate list exceeds bounded exact ranking limit")
	}
	out := make([]Scored, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Completeness != embeddings.Ready {
			continue
		}
		if err := candidate.Validate(); err != nil || candidate.ProfileFingerprint != profile.Fingerprint || candidate.Dimension != profile.Dimension || candidate.Distance != profile.Distance {
			return nil, errors.New("candidate does not match embedding profile")
		}
		score, err := embeddings.Cosine(query, candidate.Vector)
		if err != nil {
			return nil, err
		}
		out = append(out, Scored{Record: candidate, Score: score})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Record.ID < out[j].Record.ID
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// Scored is an exact cosine result from a bounded candidate list.
type Scored struct {
	// Record is the candidate whose vector was exactly compared to the query.
	Record embeddings.Record
	// Score is cosine similarity in the inclusive range [-1, 1].
	Score float32
}
