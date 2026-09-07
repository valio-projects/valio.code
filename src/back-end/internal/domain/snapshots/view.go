package snapshots

import (
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/projects"
)

// View pins both project definitions and file memberships. Changing a project
// never changes the meaning of a previously published view.
type View struct {
	// ID identifies this immutable record or registered catalog identity.
	ID string `json:"id"`
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// SnapshotID pins the concrete source snapshot used by this record.
	SnapshotID string `json:"snapshotId"`
	// Repositories pins the selected snapshot and manifest for every included repository.
	Repositories []domain.RepositorySnapshot `json:"repositories"`
	// Projects preserves complete definitions as they existed at view publication.
	Projects []projects.Definition `json:"projects"`
	// ProjectRevisions links pinned project definitions to immutable canonical fingerprints.
	ProjectRevisions []domain.ProjectRevisionRef `json:"projectRevisions"`
	// Files preserves source references and their project memberships within this view.
	Files []FileRef `json:"files"`
	// Mixed explicitly marks a view that combines multiple repository snapshots.
	Mixed bool `json:"mixed"`
	// Status reports actual completeness or command outcome without implying compiler evidence.
	Status string `json:"status"`
	// Projections reports ready, partial or unsupported for each projection family.
	Projections map[string]string `json:"projections"`
	// Profile names the actual syntax producer and profile namespace.
	Profile string `json:"profile"`
}
