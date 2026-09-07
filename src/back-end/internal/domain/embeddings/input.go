package embeddings

import (
	"errors"
	"strings"

	"github.com/valio-projects/valio.code/internal/domain"
)

// Input is a policy-approved representation supplied for a single workspace/view computation.
// Representation is intentionally transient and is never persisted by this model.
type Input struct {
	// ID is a stable caller-assigned identity within the workspace/view.
	ID string `json:"id"`
	// WorkspaceID scopes the input to one workspace.
	WorkspaceID domain.WorkspaceID `json:"workspaceId"`
	// ViewID scopes the input to one immutable view.
	ViewID domain.ViewID `json:"viewId"`
	// Kind identifies the representation class.
	Kind Kind `json:"kind"`
	// Fingerprint is the SHA-256 digest of Representation.
	Fingerprint string `json:"fingerprint"`
	// Representation is policy-approved model input and is not serialized or retained.
	Representation string `json:"-"`
	// PolicyApproved confirms the caller evaluated source-upload policy before invoking an embedder.
	PolicyApproved bool `json:"-"`
}

// Validate reports whether i is bounded, policy-approved, and bound to its representation fingerprint.
func (i Input) Validate() error {
	if i.ID == "" || strings.TrimSpace(i.ID) != i.ID || i.WorkspaceID == "" || i.ViewID == "" ||
		!i.Kind.IsValid() || i.Fingerprint != Hash(i.Representation) || !i.PolicyApproved ||
		i.Representation == "" || len(i.Representation) > 1<<20 {
		return errors.New("invalid embedding input")
	}
	return nil
}
