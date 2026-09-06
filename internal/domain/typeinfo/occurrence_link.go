package typeinfo

import (
	"github.com/valio-projects/valio.code/internal/domain"
)

// OccurrenceLink joins metadata to a source occurrence and optional symbol.
type OccurrenceLink struct {
	// ID identifies the occurrence in its analysis result.
	ID string `json:"id"`
	// Range is the exact half-open source span.
	Range domain.SourceRange `json:"range"`
	// Role describes the occurrence's producer-defined role.
	Role Fact[string] `json:"role"`
	// Symbol preserves resolution state instead of guessing a target.
	Symbol SymbolReference `json:"symbol"`
}
