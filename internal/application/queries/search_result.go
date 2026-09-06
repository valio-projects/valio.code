package queries

import "github.com/valio-projects/valio.code/internal/search"

// SearchResult returns verified matches and the concrete view used for the query.
type SearchResult struct {
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
	search.Result
}
