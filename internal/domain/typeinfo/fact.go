package typeinfo

import (
	"encoding/json"
	"fmt"
)

// Fact never represents unknown as a zero value. Its scope is inherited from the
// owning descriptor only when Scope is empty; known facts require explicit scope
// and evidence. Zero-value facts serialize as unresolved, never as known zero.
type Fact[T any] struct {
	Status   FactStatus    `json:"status"`
	Value    *T            `json:"value"`
	Scope    TypeScope     `json:"scope"`
	Evidence []EvidenceRef `json:"evidence"`
	Reason   string        `json:"reason,omitempty"`
}

func KnownFact[T any](value T, scope TypeScope, evidence ...EvidenceRef) Fact[T] {
	return Fact[T]{Status: FactKnown, Value: &value, Scope: scope, Evidence: append([]EvidenceRef{}, evidence...)}
}
func UnresolvedFact[T any](scope TypeScope, reason string) Fact[T] {
	return Fact[T]{Status: FactUnresolved, Scope: scope, Evidence: []EvidenceRef{}, Reason: reason}
}
func UnsupportedFact[T any](scope TypeScope, reason string) Fact[T] {
	return Fact[T]{Status: FactUnsupported, Scope: scope, Evidence: []EvidenceRef{}, Reason: reason}
}
func (f Fact[T]) EffectiveStatus() FactStatus {
	if f.Status == "" {
		return FactUnresolved
	}
	return f.Status
}
func (f Fact[T]) Validate(owner TypeScope) error {
	status := f.EffectiveStatus()
	if status != FactKnown && status != FactUnresolved && status != FactUnsupported {
		return fmt.Errorf("invalid fact status %q", status)
	}
	if f.Scope != (TypeScope{}) && f.Scope != owner {
		return fmt.Errorf("fact scope differs from owning type")
	}
	if status == FactKnown {
		if f.Value == nil || f.Scope != owner || len(f.Evidence) == 0 {
			return fmt.Errorf("known fact requires value, explicit owner scope and evidence")
		}
	} else if f.Value != nil {
		return fmt.Errorf("unknown/unsupported fact cannot carry a value")
	}
	for _, e := range f.Evidence {
		if e.ID == "" || e.Producer == "" {
			return fmt.Errorf("fact evidence requires identity and producer")
		}
		if e.Range != nil {
			if err := e.Range.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}
func (f Fact[T]) MarshalJSON() ([]byte, error) {
	type encoded Fact[T]
	copy := encoded(f)
	copy.Status = f.EffectiveStatus()
	if copy.Evidence == nil {
		copy.Evidence = []EvidenceRef{}
	}
	if copy.Status != FactKnown && copy.Reason == "" {
		copy.Reason = "not collected"
	}
	return json.Marshal(copy)
}
