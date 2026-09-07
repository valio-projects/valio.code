package catalog

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain"
)

// RepositoryCommandHandler dispatches repository registration through catalog validation.
type RepositoryCommandHandler struct { // Service supplies the application behavior invoked by this handler.
	Service *Service
}

// Handle dispatches its command or query through the injected service; ctx controls cancellation.
func (h RepositoryCommandHandler) Handle(ctx context.Context, c RegisterRepositoryCommand) (domain.Repository, error) {
	return h.Service.Register(ctx, c.Repository)
}
