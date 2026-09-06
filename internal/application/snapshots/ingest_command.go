package snapshots

import (
	"context"
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/domain"
)

// IngestCommand accepts a sanitized snapshot for a registered repository and workspace.
type IngestCommand struct {
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// RepositoryID identifies the explicitly registered source repository.
	RepositoryID domain.RepositoryID `json:"repositoryId"`
	// Snapshot carries validated agent input or separately stored snapshot metadata.
	Snapshot agent.Snapshot `json:"snapshot"`
}

// IngestHandler defines transport-independent atomic upload dispatch.
type IngestHandler interface {
	// Handle dispatches its command or query through the injected service; ctx controls cancellation.
	Handle(context.Context, IngestCommand) (IngestResult, error)
}
