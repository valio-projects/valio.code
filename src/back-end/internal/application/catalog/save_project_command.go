package catalog

import (
	"context"
	"github.com/valio-projects/valio.code/internal/projects"
)

// SaveProjectCommand requests creation or validated replacement of a project definition.
type SaveProjectCommand struct {
	// Definition contains the complete project metadata and source-root definition.
	Definition projects.Definition
	// Update requires an existing project and preserves its identity and stable key.
	Update bool
}

// SaveProjectHandler defines transport-independent project command dispatch.
type SaveProjectHandler interface {
	// Handle dispatches its command or query through the injected service; ctx controls cancellation.
	Handle(context.Context, SaveProjectCommand) (projects.Definition, error)
}
