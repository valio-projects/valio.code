package repositories

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/projects"
)

// CatalogRepository stores registered repositories and current project
// definitions. Implementations enforce immutable keys and compare old writes.
type CatalogRepository interface {
	// Projects returns all bounded current definitions in the configured workspace; ctx controls cancellation.
	Projects(context.Context) ([]projects.Definition, error)
	// Project finds id only within the configured workspace; ctx controls cancellation.
	Project(context.Context, string) (projects.Definition, error)
	// StoreProject persists d after comparing old when updating; ctx controls cancellation.
	StoreProject(context.Context, projects.Definition, *projects.Definition) error
	// Repositories lists bounded registered repository identities; ctx controls cancellation.
	Repositories(context.Context) ([]domain.Repository, error)
	// Repository finds a registered repository by id in the configured workspace; ctx controls cancellation.
	Repository(context.Context, string) (domain.Repository, error)
	// RegisterRepository creates r with immutable identity; ctx controls cancellation.
	RegisterRepository(context.Context, domain.Repository) error
}
