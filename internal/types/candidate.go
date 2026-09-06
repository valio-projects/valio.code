package types

import (
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// Candidate pairs a matched descriptor with counted records and resolution evidence.
type Candidate struct {
	// Type is a deep-copied descriptor matching the query.
	Type typeinfo.TypeDescriptor `json:"type"`
	// Counts covers recorded metadata, not all possible semantic members.
	Counts MemberCounts `json:"counts"`
	// Reference is exact for each candidate even when the overall query is ambiguous.
	Reference typeinfo.SymbolReference `json:"reference"`
}
