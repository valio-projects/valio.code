package queries

import (
	"context"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/retrieval"
)

// Context selects canonical snippets with citations and explicit omitted items.
func (s Service) Context(ctx context.Context, q ContextQuery) (retrieval.Context, error) {
	if q.ChunkID == "" || q.MaxBytes < 0 || q.MaxBytes > 8000 {
		return retrieval.Context{}, fault.ErrInvalid
	}
	v, chunks, err := s.scopedChunks(ctx, q.Scope)
	if err != nil {
		return retrieval.Context{}, err
	}
	result, err := (retrieval.ContextBuilder{}).Build(q.ChunkID, chunks, retrieval.ContextOptions{ViewID: v.ID, ProjectIDs: q.Scope.ProjectIDs, MaxBytes: q.MaxBytes})
	if err != nil {
		return result, fault.ErrInvalid
	}
	return result, nil
}
