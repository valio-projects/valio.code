package typeinfo

import (
	"encoding/json"
	"fmt"
)

// Fact never represents unknown as a zero value. Its scope is inherited from the
// owning descriptor only when Scope is empty; known facts require explicit scope
// and evidence. Zero-value facts serialize as unresolved, never as known zero.
type Fact[T any] struct {
	// Status distinguishes known values from unresolved and unsupported facts.
	Status FactStatus `json:"status"`
	// Value is present only when Status is known.
	Value *T `json:"value"`
	// Scope explicitly identifies the input that established a known value.
	Scope TypeScope `json:"scope"`
	// Evidence identifies producers and optional source spans supporting the fact.
	Evidence []EvidenceRef `json:"evidence"`
	// Reason explains unresolved or unsupported status when supplied.
	Reason string `json:"reason,omitempty"`
}

// KnownFact creates a scoped fact from value and its nonempty supporting evidence.
func KnownFact[T any](value T, scope TypeScope, evidence ...EvidenceRef) Fact[T] {
	return Fact[T]{Status: FactKnown, Value: &value, Scope: scope, Evidence: append([]EvidenceRef{}, evidence...)}
}

// UnresolvedFact records an expected value that the producer could not establish.
func UnresolvedFact[T any](scope TypeScope, reason string) Fact[T] {
	return Fact[T]{Status: FactUnresolved, Scope: scope, Evidence: []EvidenceRef{}, Reason: reason}
}

// UnsupportedFact records a value the active producer cannot collect.
func UnsupportedFact[T any](scope TypeScope, reason string) Fact[T] {
	return Fact[T]{Status: FactUnsupported, Scope: scope, Evidence: []EvidenceRef{}, Reason: reason}
}

// EffectiveStatus maps an omitted status to unresolved for backward-safe decoding.
func (f Fact[T]) EffectiveStatus() FactStatus {
	if f.Status == "" {
		return FactUnresolved
	}
	return f.Status
}

// Validate checks that fact payload, scope, and evidence agree with owner.
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

// MarshalJSON serializes zero-value facts as unresolved and keeps evidence arrays non-null.
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
