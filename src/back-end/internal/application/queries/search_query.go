package queries

import (
	"github.com/valio-projects/valio.code/internal/search"
)

// SearchQuery requests a bounded search over one immutable view.
type SearchQuery struct {
	// Query contains the bounded boolean source-search expression.
	Query string `json:"query"`
	// Mode selects substring, exact or regular-expression matching.
	Mode search.Mode `json:"mode,omitempty"`
	// Scope selects the workspace, pinned view and optional projects.
	Scope SearchScope `json:"scope"`
	// Limit bounds returned matches after the complete scoped evaluation.
	Limit int `json:"limit,omitempty"`
	// Offset skips this many matches after deterministic ordering.
	Offset int `json:"offset,omitempty"`
	// CaseSensitive preserves source case during matching when true.
	CaseSensitive bool `json:"caseSensitive,omitempty"`
}
