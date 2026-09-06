package types

import (
	"github.com/valio-projects/valio.code/internal/domain"
)

type QueryScope struct {
	WorkspaceID    domain.WorkspaceID `json:"workspaceId"`
	ProjectID      domain.ProjectID   `json:"projectId,omitempty"`
	BuildProfileID string             `json:"buildProfileId,omitempty"`
	VersionID      string             `json:"versionId,omitempty"`
}
