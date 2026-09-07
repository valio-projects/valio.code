// Package embeddings defines immutable, view-scoped dense-vector contracts.
package embeddings

// Kind identifies an independently versioned representation supplied to a model.
type Kind string

const (
	// Code represents canonical source-code content.
	Code Kind = "code"
	// Symbol represents a declaration or exported-symbol representation.
	Symbol Kind = "symbol"
	// Context represents selected bounded code context.
	Context Kind = "context"
	// Documentation represents documentation associated with code.
	Documentation Kind = "documentation"
	// Architecture represents a structural or architectural representation.
	Architecture Kind = "architecture"
	// Change represents a change-oriented representation.
	Change Kind = "change"
	// Error represents an error or diagnostic representation.
	Error Kind = "error"
	// API represents an API-contract representation.
	API Kind = "api"
)

// Valid reports whether k is one of the eight supported representation kinds.
func (k Kind) Valid() bool {
	switch k {
	case Code, Symbol, Context, Documentation, Architecture, Change, Error, API:
		return true
	}
	return false
}

// IsValid reports whether k is one of the eight supported representation kinds.
func (k Kind) IsValid() bool { return k.Valid() }
