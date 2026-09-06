package types

import (
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

type Candidate struct {
	Type      typeinfo.TypeDescriptor  `json:"type"`
	Counts    MemberCounts             `json:"counts"`
	Reference typeinfo.SymbolReference `json:"reference"`
}
