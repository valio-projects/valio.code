package types

import (
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

type NameResult struct {
	Status     QueryStatus              `json:"status"`
	Candidates []Candidate              `json:"candidates"`
	Reference  typeinfo.SymbolReference `json:"reference"`
}
