package types

import (
	"github.com/valio-projects/valio.code/internal/domain"
)

// QueryScope narrows a lookup; only WorkspaceID is required by name queries.
type QueryScope struct {
	// WorkspaceID is the required identity boundary.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// ProjectID optionally narrows results to one project.
	ProjectID domain.ProjectID `json:"projectId,omitempty"`
	// BuildProfileID optionally narrows results to one build profile.
	BuildProfileID string `json:"buildProfileId,omitempty"`
	// VersionID optionally narrows results to one immutable analysis version.
	VersionID string `json:"versionId,omitempty"`
}
