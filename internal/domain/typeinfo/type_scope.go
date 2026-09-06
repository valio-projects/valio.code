package typeinfo

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain"
)

// TypeScope pins a type fact to a workspace, project, build profile and immutable
// analysis version. Identically named types in another scope are distinct.
type TypeScope struct {
	WorkspaceID    domain.WorkspaceID `json:"workspaceId"`
	ProjectID      domain.ProjectID   `json:"projectId"`
	BuildProfileID string             `json:"buildProfileId"`
	VersionID      string             `json:"versionId"`
}

func (s TypeScope) Validate() error {
	if s.WorkspaceID == "" || s.ProjectID == "" || s.BuildProfileID == "" || s.VersionID == "" {
		return fmt.Errorf("type scope requires workspace, project, build profile and version")
	}
	return nil
}
