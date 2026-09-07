package catalog

import (
	"context"
	"github.com/valio-projects/valio.code/internal/projects"
)

// ProjectCommandHandler dispatches project writes through catalog validation.
type ProjectCommandHandler struct { // Service supplies the application behavior invoked by this handler.
	Service *Service
}

// Handle dispatches its command or query through the injected service; ctx controls cancellation.
func (h ProjectCommandHandler) Handle(ctx context.Context, c SaveProjectCommand) (projects.Definition, error) {
	return h.Service.SaveProject(ctx, c.Definition, c.Update)
}
