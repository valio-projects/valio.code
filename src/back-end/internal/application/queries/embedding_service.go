package queries

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	appembedding "github.com/valio-projects/valio.code/internal/application/embeddings"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/embeddings"
	"sort"
)

// IndexEmbeddings computes only a bounded page, preserving completed immutable
// records if a later provider request fails. A retry reuses those records.
func (s Service) IndexEmbeddings(ctx context.Context, q EmbeddingCommand) (EmbeddingResult, error) {
	result := EmbeddingResult{}
	if q.Offset < 0 || q.Offset > 100000 || q.Limit < 0 || q.Limit > 16 {
		return result, fault.ErrInvalid
	}
	if q.Limit == 0 {
		q.Limit = 8
	}
	kind := embeddings.Kind(q.Representation)
	if kind == "" {
		kind = embeddings.Code
	}
	if !kind.IsValid() {
		return result, fault.ErrInvalid
	}
	if s.Models == nil || s.Vectors == nil {
		return result, fault.ErrUnavailable
	}
	profile, provider, err := s.Models.ResolveModel(q.ModelProfile, kind)
	if err != nil || provider == nil {
		return result, fault.ErrUnavailable
	}
	v, chunks, err := s.scopedChunks(ctx, q.Scope)
	if err != nil {
		return result, err
	}
	inputs := []embeddings.Input{}
	for _, c := range chunks {
		for _, r := range c.Representations {
			if string(r.Kind) != string(kind) || r.Text == "" {
				continue
			}
			sum := sha256.Sum256([]byte(r.Text))
			inputs = append(inputs, embeddings.Input{ID: c.ID, WorkspaceID: s.WorkspaceID, ViewID: domain.ViewID(v.ID), Kind: kind, Fingerprint: hex.EncodeToString(sum[:]), Representation: r.Text, PolicyApproved: true})
			break
		}
	}
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].ID < inputs[j].ID })
	result.ViewID = v.ID
	result.ProfileFingerprint = profile.Fingerprint
	result.Eligible = len(inputs)
	result.NextOffset = min(q.Offset, len(inputs))
	service := appembedding.Service{Repository: s.Vectors, Provider: provider}
	for i := min(q.Offset, len(inputs)); i < min(q.Offset+q.Limit, len(inputs)); i++ {
		if _, err := service.Compute(ctx, profile, inputs[i]); err != nil {
			if ctx.Err() != nil {
				return result, ctx.Err()
			}
			return result, fault.ErrUnavailable
		}
		result.Processed++
		result.NextOffset = i + 1
	}
	result.Complete = result.NextOffset == len(inputs)
	return result, nil
}
