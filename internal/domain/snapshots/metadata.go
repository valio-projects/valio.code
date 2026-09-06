package snapshots

import (
	"github.com/valio-projects/valio.code/internal/agent"
	"github.com/valio-projects/valio.code/internal/configgraph"
	"github.com/valio-projects/valio.code/internal/domain"
	gitrepo "github.com/valio-projects/valio.code/internal/git"
)

// Metadata retains snapshot provenance and sanitized configuration metadata;
// source bytes and manifests are separately stored, immutable records.
type Metadata struct {
	// ID identifies this immutable record or registered catalog identity.
	ID string `json:"id"`
	// AgentSnapshotID preserves the validated agent payload identity before workspace scoping.
	AgentSnapshotID string `json:"agentSnapshotId"`
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// RepositoryID identifies the explicitly registered source repository.
	RepositoryID domain.RepositoryID `json:"repositoryId"`
	// Repository carries registered identity or sanitized Git provenance, never executable paths.
	Repository gitrepo.Repository `json:"repository"`
	// Config retains only filtered configuration-key metadata, excluding values.
	Config []configgraph.Projection `json:"config"`
	// Diagnostics retains sanitized capture outcomes without source or credential payloads.
	Diagnostics []agent.Diagnostic `json:"diagnostics"`
	// ManifestFingerprint pins the complete sorted repository content manifest.
	ManifestFingerprint string `json:"manifestFingerprint"`
}
