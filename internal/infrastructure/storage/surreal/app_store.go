package surreal

import (
	"context"
	"errors"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/fault"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/projects"
)

// AppStore is permanently bound to one database and one bootstrap workspace.
type AppStore struct {
	client       *Client
	workspace    domain.WorkspaceID
	projects     *Repository[projects.Definition]
	repositories *Repository[domain.Repository]
	views        *Repository[snapshots.View]
}

// NewAppStore binds client and workspace permanently; it does not change database credentials or scope.
func NewAppStore(client *Client, workspace domain.WorkspaceID) (*AppStore, error) {
	if client == nil || workspace == "" {
		return nil, errors.New("database and workspace required")
	}
	p, e := NewRepository[projects.Definition](client, "project", false)
	if e != nil {
		return nil, e
	}
	r, e := NewRepository[domain.Repository](client, "repository", true)
	if e != nil {
		return nil, e
	}
	v, e := NewRepository[snapshots.View](client, "analysis_view", true)
	if e != nil {
		return nil, e
	}
	return &AppStore{client: client, workspace: workspace, projects: p, repositories: r, views: v}, nil
}
func appError(e error) error {
	if errors.Is(e, ErrNotFound) {
		return fault.ErrNotFound
	}
	return e
}

// Projects returns all bounded current definitions in the configured workspace; ctx controls cancellation.
func (s *AppStore) Projects(ctx context.Context) ([]projects.Definition, error) {
	values, e := s.projects.List(ctx, 1000, 0)
	if e != nil {
		return nil, e
	}
	if len(values) >= 1000 {
		return nil, fault.ErrScopeTooLarge
	}
	for _, v := range values {
		if v.Project.WorkspaceID != s.workspace {
			return nil, fault.ErrForbidden
		}
	}
	return values, nil
}

// Project finds id only within the configured workspace; ctx controls cancellation.
func (s *AppStore) Project(ctx context.Context, id string) (projects.Definition, error) {
	v, e := s.projects.Find(ctx, id)
	if e != nil {
		return v, appError(e)
	}
	if v.Project.WorkspaceID != s.workspace {
		return v, fault.ErrForbidden
	}
	return v, nil
}

// Repositories lists bounded registered repository identities; ctx controls cancellation.
func (s *AppStore) Repositories(ctx context.Context) ([]domain.Repository, error) {
	values, e := s.repositories.List(ctx, 1000, 0)
	if e != nil {
		return nil, e
	}
	if len(values) >= 1000 {
		return nil, fault.ErrScopeTooLarge
	}
	for _, v := range values {
		if v.WorkspaceID != s.workspace {
			return nil, fault.ErrForbidden
		}
	}
	return values, nil
}

// Repository finds a registered repository by id in the configured workspace; ctx controls cancellation.
func (s *AppStore) Repository(ctx context.Context, id string) (domain.Repository, error) {
	v, e := s.repositories.Find(ctx, id)
	if e != nil {
		return v, appError(e)
	}
	if v.WorkspaceID != s.workspace {
		return v, fault.ErrForbidden
	}
	return v, nil
}

// RegisterRepository creates r with immutable identity; ctx controls cancellation.
func (s *AppStore) RegisterRepository(ctx context.Context, r domain.Repository) error {
	if r.WorkspaceID != s.workspace {
		return fault.ErrForbidden
	}
	if e := s.repositories.Create(ctx, string(r.ID), r); e != nil {
		return fault.ErrConflict
	}
	return nil
}

// StoreProject persists d after comparing old when updating; ctx controls cancellation.
func (s *AppStore) StoreProject(ctx context.Context, d projects.Definition, old *projects.Definition) error {
	if d.Project.WorkspaceID != s.workspace {
		return fault.ErrForbidden
	}
	sql := "BEGIN TRANSACTION; LET $existing = (SELECT VALUE payload FROM type::record('project',$id)); IF array::len($existing) != 0 { THROW 'conflict'; }; CREATE type::record('project',$id) SET key=$id,payload=$definition; COMMIT TRANSACTION;"
	vars := map[string]any{"id": string(d.Project.ID), "definition": d}
	if old != nil {
		sql = "BEGIN TRANSACTION; LET $existing = (SELECT VALUE payload FROM type::record('project',$id)); IF array::len($existing) != 1 OR $existing[0] != $previous { THROW 'conflict'; }; UPDATE type::record('project',$id) SET payload=$definition; COMMIT TRANSACTION;"
		vars["previous"] = old
	}
	_, e := s.client.Query(ctx, sql, vars)
	if e != nil {
		return fault.ErrConflict
	}
	return nil
}

// View reads id as an immutable view and verifies workspace scope; ctx controls cancellation.
func (s *AppStore) View(ctx context.Context, id string) (snapshots.View, error) {
	v, e := s.views.Find(ctx, id)
	if e != nil {
		return v, appError(e)
	}
	if v.WorkspaceID != s.workspace {
		return v, fault.ErrForbidden
	}
	return v, nil
}

// Latest resolves the current publication pointer into one immutable view; ctx controls cancellation.
func (s *AppStore) Latest(ctx context.Context) (snapshots.View, error) {
	head, e := Get[string](ctx, s.client, "projection_head", "app-latest")
	if e != nil {
		return snapshots.View{}, appError(e)
	}
	return s.View(ctx, head)
}

// Bootstrap refuses to bind a database already owned by another workspace.
func (s *AppStore) Bootstrap(ctx context.Context, w domain.Workspace) error {
	if w.ID != s.workspace {
		return fault.ErrForbidden
	}
	_, e := s.client.Query(ctx, `BEGIN TRANSACTION; LET $existing=(SELECT VALUE payload FROM workspace:bootstrap); IF array::len($existing)=0 { CREATE workspace:bootstrap SET key='bootstrap',payload=$workspace; } ELSE IF $existing[0].id != $workspace.id { THROW 'workspace conflict'; }; COMMIT TRANSACTION;`, map[string]any{"workspace": w})
	return e
}
