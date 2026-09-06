package catalog

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain"
)

// RegisterRepositoryCommand registers an explicit repository identity before upload.
type RegisterRepositoryCommand struct { // Repository carries registered identity or sanitized Git provenance, never executable paths.
	Repository domain.Repository
}

// RegisterRepositoryHandler defines transport-independent repository command dispatch.
type RegisterRepositoryHandler interface {
	// Handle dispatches its command or query through the injected service; ctx controls cancellation.
	Handle(context.Context, RegisterRepositoryCommand) (domain.Repository, error)
}
