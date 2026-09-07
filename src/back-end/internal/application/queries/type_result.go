package queries

import "github.com/valio-projects/valio.code/internal/types"

// TypeResult returns scoped rich type candidates without merging ambiguous identities.
type TypeResult struct {
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
	types.NameResult
}
