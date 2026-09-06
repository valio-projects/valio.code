package search

// Mode selects matching semantics for text terms.
type Mode string

const (
	// Substring matches a term anywhere in the searched value.
	Substring Mode = "substring"
	// Exact matches an entire searched value.
	Exact Mode = "exact"
	// Regex evaluates a Go RE2 expression.
	Regex Mode = "regex"
)
