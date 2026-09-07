package typeinfo

import (
	"github.com/valio-projects/valio.code/internal/domain"
)

// EvidenceRef identifies the producer or source that established a fact. Range
// is optional for compiler/layout evidence with no single source span.
type EvidenceRef struct {
	ID       string              `json:"id"`
	Producer string              `json:"producer"`
	Range    *domain.SourceRange `json:"range,omitempty"`
}
