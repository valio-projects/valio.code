package projects

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain"
)

// ManifestBuilder binds all built manifests to one workspace and repository.
type ManifestBuilder struct {
	workspace  domain.WorkspaceID
	repository domain.RepositoryID
}

// NewManifestBuilder creates a builder fixed to nonempty workspace and repository IDs.
func NewManifestBuilder(workspace domain.WorkspaceID, repository domain.RepositoryID) (*ManifestBuilder, error) {
	if workspace == "" || repository == "" {
		return nil, fmt.Errorf("manifest builder requires workspace and repository")
	}
	return &ManifestBuilder{workspace: workspace, repository: repository}, nil
}
