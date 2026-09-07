package queries

import (
	"context"
	"fmt"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"slices"
)

// scopedChunks verifies the immutable artifact chain before returning any text.
func (s Service) scopedChunks(ctx context.Context, scope SearchScope) (snapshots.View, []retrieval.Chunk, error) {
	if scope.WorkspaceID != s.WorkspaceID {
		return snapshots.View{}, nil, fault.ErrForbidden
	}
	if len(scope.ProjectIDs) > 128 {
		return snapshots.View{}, nil, fault.ErrInvalid
	}
	v, err := s.View(ctx, scope.ViewID)
	if err != nil {
		return v, nil, err
	}
	for _, id := range scope.ProjectIDs {
		found := false
		for _, p := range v.Projects {
			if string(p.Project.ID) == id {
				found = true
				break
			}
		}
		if !found {
			return v, nil, fault.ErrNotFound
		}
	}
	if v.Projections["retrieval_chunks"] != "ready" {
		return v, nil, fault.ErrUnavailable
	}
	artifacts, err := s.Store.Artifacts(ctx, v)
	if err != nil {
		return v, nil, err
	}
	files := map[string]snapshots.FileRef{}
	for _, f := range v.Files {
		files[f.ID] = f
	}
	output := []retrieval.Chunk{}
	seen := map[string]bool{}
	bytes := 0
	for _, a := range artifacts {
		if err := ctx.Err(); err != nil {
			return v, nil, err
		}
		f, ok := files[a.FileID]
		if !ok || a.ViewID != v.ID {
			return v, nil, fault.ErrForbidden
		}
		for _, chunk := range a.Chunks {
			if chunk.ID == "" || seen[chunk.ID] {
				return v, nil, fmt.Errorf("chunk identity is empty or duplicated: %w", fault.ErrInvalid)
			}
			if chunk.FileID != f.ID || chunk.RepositoryID != string(f.RepositoryID) || !slices.Equal(chunk.ProjectIDs, f.ProjectIDs) {
				return v, nil, fmt.Errorf("chunk membership differs from source: %w", fault.ErrInvalid)
			}
			if chunk.Start < 0 || chunk.End < chunk.Start || chunk.End > f.Size || chunk.Text != "" && len(chunk.Text) != chunk.End-chunk.Start {
				return v, nil, fmt.Errorf("chunk range differs from source: %w", fault.ErrInvalid)
			}
			seen[chunk.ID] = true
			if len(scope.ProjectIDs) > 0 && !overlaps(chunk.ProjectIDs, scope.ProjectIDs) {
				continue
			}
			bytes += len(chunk.Text) + len(chunk.Header)
			if len(output) >= 100000 || bytes > 32<<20 {
				return v, nil, fault.ErrScopeTooLarge
			}
			output = append(output, chunk)
		}
	}
	return v, output, nil
}

func overlaps(a, b []string) bool {
	for _, id := range a {
		if slices.Contains(b, id) {
			return true
		}
	}
	return false
}
