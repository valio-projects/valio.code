package queries

import (
	"context"
	"errors"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
	"github.com/valio-projects/valio.code/internal/search"
	"github.com/valio-projects/valio.code/internal/types"
)

// Service resolves a mutable latest pointer once, then uses that pinned view
// throughout the query. It never recomputes membership from current projects.
type Service struct {
	// Store supplies the typed, workspace-bound persistence port.
	Store repositories.SnapshotRepository
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID
}

// View resolves id, or the current latest pointer exactly once when id is empty.
func (s Service) View(ctx context.Context, id string) (snapshots.View, error) {
	var v snapshots.View
	var e error
	if id == "" {
		v, e = s.Store.Latest(ctx)
	} else {
		v, e = s.Store.View(ctx, id)
	}
	if e == nil && v.WorkspaceID != s.WorkspaceID {
		return v, fault.ErrForbidden
	}
	return v, e
}

// Search verifies all scoped files before paging; exhausted budgets are errors.
func (s Service) Search(ctx context.Context, q SearchQuery) (SearchResult, error) {
	result := SearchResult{}
	if q.Scope.WorkspaceID != s.WorkspaceID {
		return result, fault.ErrForbidden
	}
	if q.Query == "" || len(q.Query) > 8192 || q.Limit < 0 || q.Limit > 1000 || q.Offset < 0 || len(q.Scope.ProjectIDs) > 128 {
		return result, fault.ErrInvalid
	}
	if q.Mode != "" && q.Mode != search.Exact && q.Mode != search.Substring && q.Mode != search.Regex {
		return result, fault.ErrInvalid
	}
	parsed, e := search.Parse(q.Query)
	if e != nil {
		return result, fault.ErrInvalid
	}
	v, e := s.View(ctx, q.Scope.ViewID)
	if e != nil {
		return result, e
	}
	for _, id := range q.Scope.ProjectIDs {
		found := false
		for _, d := range v.Projects {
			if string(d.Project.ID) == id {
				found = true
			}
		}
		if !found {
			return result, fault.ErrNotFound
		}
	}
	files, e := s.Store.Files(ctx, v)
	if e != nil {
		return result, e
	}
	matches, e := search.Execute(ctx, search.MemorySource(files), parsed, search.Options{ProjectIDs: q.Scope.ProjectIDs, Mode: q.Mode, CaseSensitive: q.CaseSensitive, Limit: q.Limit, Offset: q.Offset, MaxScanFiles: 10000, MaxScanBytes: 16 << 20})
	if e != nil {
		if errors.Is(e, context.Canceled) || errors.Is(e, context.DeadlineExceeded) {
			return result, e
		}
		return result, fault.ErrInvalid
	}
	if !matches.Complete {
		return result, fault.ErrScopeTooLarge
	}
	return SearchResult{ViewID: v.ID, Result: matches}, nil
}

// Types matches exact simple or qualified names only within the selected view.
func (s Service) Types(ctx context.Context, q TypeQuery) (TypeResult, error) {
	result := TypeResult{}
	if q.Name == "" || len(q.Name) > 4096 {
		return result, fault.ErrInvalid
	}
	v, e := s.View(ctx, q.ViewID)
	if e != nil {
		return result, e
	}
	artifacts, e := s.Store.Artifacts(ctx, v)
	if e != nil {
		return result, e
	}
	descriptors := []typeinfo.TypeDescriptor{}
	for _, a := range artifacts {
		for _, d := range a.Types {
			if d.Scope.WorkspaceID != s.WorkspaceID || d.Scope.VersionID != v.ID {
				return result, fault.ErrForbidden
			}
			descriptors = append(descriptors, d)
		}
	}
	catalog, e := types.NewCatalog(descriptors)
	if e != nil {
		return result, e
	}
	found, e := catalog.ByName(types.QueryScope{WorkspaceID: s.WorkspaceID, ProjectID: domain.ProjectID(q.ProjectID), BuildProfileID: q.BuildProfileID, VersionID: v.ID}, q.Name)
	return TypeResult{ViewID: v.ID, NameResult: found}, e
}

// File returns source only if fileID is a member of the selected immutable view.
func (s Service) File(ctx context.Context, viewID, fileID string) (search.File, error) {
	v, e := s.View(ctx, viewID)
	if e != nil {
		return search.File{}, e
	}
	for _, f := range v.Files {
		if f.ID == fileID {
			v.Files = []snapshots.FileRef{f}
			files, e := s.Store.Files(ctx, v)
			if e != nil {
				return search.File{}, e
			}
			if len(files) == 1 {
				return files[0], nil
			}
		}
	}
	return search.File{}, fault.ErrNotFound
}
