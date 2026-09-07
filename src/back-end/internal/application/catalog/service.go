package catalog

import (
	"context"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/projects"
	"github.com/valio-projects/valio.code/internal/validation"
	"net/url"
	"regexp"
	"strings"
)

// Service validates project definitions and explicit repository registrations.
type Service struct {
	// Store supplies the typed, workspace-bound persistence port.
	Store Repository
	// Workspace defines the only workspace authorized by this bootstrap service.
	Workspace domain.Workspace
}

var identity = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)

// ValidID accepts portable caller-assigned catalog identifiers.
func ValidID(id string) bool { return identity.MatchString(id) }

// SaveProject validates d and preserves immutable identity on update; ctx bounds repository work.
func (s Service) SaveProject(ctx context.Context, d projects.Definition, update bool) (projects.Definition, error) {
	if d.Project.WorkspaceID != s.Workspace.ID {
		return d, fault.ErrForbidden
	}
	if !ValidID(string(d.Project.ID)) || len(d.Roots) > 128 || len(d.Project.Name) > 256 || len(d.Project.Description) > 8192 || d.Validate() != nil {
		return d, fault.ErrInvalid
	}
	var previous *projects.Definition
	if update {
		old, e := s.Store.Project(ctx, string(d.Project.ID))
		if e != nil {
			return d, e
		}
		if d.Project.ValidateUpdate(old.Project) != nil {
			return d, fault.ErrInvalid
		}
		previous = &old
	}
	for _, root := range d.Roots {
		r, e := s.Store.Repository(ctx, string(root.RepositoryID))
		if e != nil {
			return d, e
		}
		if r.WorkspaceID != s.Workspace.ID {
			return d, fault.ErrForbidden
		}
	}
	return d, s.Store.StoreProject(ctx, d, previous)
}

// Register validates r before creating an explicit repository identity; ctx controls cancellation.
func (s Service) Register(ctx context.Context, r domain.Repository) (domain.Repository, error) {
	if r.WorkspaceID != s.Workspace.ID {
		return r, fault.ErrForbidden
	}
	if !ValidID(string(r.ID)) || len(r.RemoteURL) > 2048 {
		return r, fault.ErrInvalid
	}
	if r.RemoteURL != "" {
		u, e := url.Parse(r.RemoteURL)
		if e != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !(u.Scheme == "https" || u.Scheme == "ssh") || strings.TrimSpace(u.Host) == "" {
			return r, fault.ErrInvalid
		}
		// Apply identical host/port/path syntax checks without contacting the remote.
		check := *u
		check.Scheme = "https"
		if _, e := validation.Endpoint(check.String(), false); e != nil {
			return r, fault.ErrInvalid
		}
	}
	return r, s.Store.RegisterRepository(ctx, r)
}
