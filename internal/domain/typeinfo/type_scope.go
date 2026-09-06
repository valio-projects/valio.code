package typeinfo

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain"
)

// TypeScope pins a type fact to a workspace, project, build profile and immutable
// analysis version. Identically named types in another scope are distinct.
type TypeScope struct {
	// WorkspaceID isolates identities between workspaces.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// ProjectID identifies the project contributing the type.
	ProjectID domain.ProjectID `json:"projectId"`
	// BuildProfileID identifies compiler options and target context.
	BuildProfileID string `json:"buildProfileId"`
	// VersionID identifies the immutable analysis or source version.
	VersionID string `json:"versionId"`
}

// Validate rejects incomplete scopes that could conflate facts from different inputs.
func (s TypeScope) Validate() error {
	if s.WorkspaceID == "" || s.ProjectID == "" || s.BuildProfileID == "" || s.VersionID == "" {
		return fmt.Errorf("type scope requires workspace, project, build profile and version")
	}
	return nil
}
