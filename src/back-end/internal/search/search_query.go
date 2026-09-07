package search

// SearchQuery is the read-request envelope; scopes and limits belong to this
// request rather than mutable shared engine state.
type SearchQuery struct {
	Expression string  `json:"expression"`
	Options    Options `json:"options"`
}
