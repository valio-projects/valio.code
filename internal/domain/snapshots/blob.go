package snapshots

import "github.com/valio-projects/valio.code/internal/domain"

// Blob stores source bytes under a workspace-scoped content-addressed identity.
type Blob struct {
	// ID identifies this immutable record or registered catalog identity.
	ID string `json:"id"`
	// WorkspaceID binds this value to the single configured workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// SHA256 records the verified digest of source content.
	SHA256 string `json:"sha256"`
	// Content contains sanitized source text; it must never be logged.
	Content string `json:"content"`
}
