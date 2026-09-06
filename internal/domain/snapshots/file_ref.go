package snapshots

import "github.com/valio-projects/valio.code/internal/domain"

// FileRef pins source content and project membership within a published view.
type FileRef struct {
	// ID identifies this immutable record or registered catalog identity.
	ID string `json:"id"`
	// RepositoryID identifies the explicitly registered source repository.
	RepositoryID domain.RepositoryID `json:"repositoryId"`
	// SnapshotID pins the concrete source snapshot used by this record.
	SnapshotID string `json:"snapshotId"`
	// Path is a portable repository-relative source path.
	Path string `json:"path"`
	// BlobID pins the workspace-scoped content-addressed source bytes.
	BlobID string `json:"blobId"`
	// Size records the expected UTF-8 source byte count.
	Size int `json:"size"`
	// Language records the source language determined from the sanitized path.
	Language string `json:"language"`
	// ProjectIDs lists the selected or pinned many-to-many project memberships.
	ProjectIDs []string `json:"projectIds"`
}
