package typeinfo

import (
	"fmt"
)

// SymbolReference preserves all candidates; a lexical name is never silently
// promoted to an exact semantic link.
type SymbolReference struct {
	Resolution ReferenceResolution `json:"resolution"`
	Exact      *SymbolIdentity     `json:"exact,omitempty"`
	Candidates []SymbolIdentity    `json:"candidates"`
	Reason     string              `json:"reason,omitempty"`
}

func (r SymbolReference) Validate() error {
	switch r.Resolution {
	case "", ReferenceUnresolved:
		if r.Exact != nil || len(r.Candidates) > 0 {
			return fmt.Errorf("unresolved reference cannot have targets")
		}
	case ReferenceExact:
		if r.Exact == nil || len(r.Candidates) > 0 {
			return fmt.Errorf("exact reference requires exactly one exact target")
		}
	case ReferenceCandidate:
		if r.Exact != nil || len(r.Candidates) == 0 {
			return fmt.Errorf("candidate reference requires candidate targets")
		}
	default:
		return fmt.Errorf("invalid reference resolution %q", r.Resolution)
	}
	targets := append([]SymbolIdentity{}, r.Candidates...)
	if r.Exact != nil {
		targets = append(targets, *r.Exact)
	}
	seen := map[SymbolIdentity]bool{}
	for _, target := range targets {
		if target.ID == "" {
			return fmt.Errorf("symbol target requires identity")
		}
		if err := target.Scope.Validate(); err != nil {
			return err
		}
		if seen[target] {
			return fmt.Errorf("duplicate symbol candidate")
		}
		seen[target] = true
	}
	return nil
}
