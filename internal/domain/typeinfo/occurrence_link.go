package typeinfo

import (
	"github.com/valio-projects/valio.code/internal/domain"
)

type OccurrenceLink struct {
	ID     string             `json:"id"`
	Range  domain.SourceRange `json:"range"`
	Role   Fact[string]       `json:"role"`
	Symbol SymbolReference    `json:"symbol"`
}
