package retrieval

import domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"

// ContextOptions constrains context assembly to one caller-selected view and
// project scope. MaxBytes defaults to and may not exceed 8000 exact UTF-8 bytes.
type ContextOptions struct {
	// ViewID identifies the immutable view represented by the caller's chunks.
	ViewID string
	// ProjectIDs further narrows eligible chunks; empty preserves the root scope.
	ProjectIDs []string
	// MaxBytes bounds the exact serialized context contribution of items.
	MaxBytes int
	// Profile identifies the requested retrieval representation.
	Profile domainretrieval.RepresentationKind
}
