package queries

import "github.com/valio-projects/valio.code/internal/domain"

// SearchScope pins an immutable view and optionally narrows its project set.
type SearchScope struct {
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// ProjectIDs lists the selected or pinned many-to-many project memberships.
	ProjectIDs []string `json:"projectIds"`
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
}
