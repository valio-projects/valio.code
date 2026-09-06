// Package projections describes versioned projection contracts and dependencies.
package projections

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type Family string

const (
	SourceText    Family = "source_text"
	SymbolType    Family = "symbol_type"
	Reference     Family = "reference"
	Call          Family = "call"
	DataFlow      Family = "data_flow"
	Structure     Family = "structure"
	SystemGraph   Family = "system_graph"
	HistoryChange Family = "history_change"
	Vector        Family = "vector"
	Context       Family = "context"
)

type Status string

const (
	NotRequested     Status = "not_requested"
	Queued           Status = "queued"
	Building         Status = "building"
	Ready            Status = "ready"
	Partial          Status = "partial"
	Stale            Status = "stale"
	Failed           Status = "failed"
	Unsupported      Status = "unsupported"
	ExcludedByPolicy Status = "excluded_by_policy"
)

func (s Status) Valid() bool {
	switch s {
	case NotRequested, Queued, Building, Ready, Partial, Stale, Failed, Unsupported, ExcludedByPolicy:
		return true
	}
	return false
}

type Contract struct {
	Family       Family   `json:"family"`
	Version      string   `json:"version"`
	Dependencies []Family `json:"dependencies"`
}
type Registry struct {
	Contracts []Contract `json:"contracts"`
}

func DefaultRegistry() Registry {
	return Registry{Contracts: []Contract{
		{SourceText, "1", nil}, {Structure, "1", []Family{SourceText}},
		{SymbolType, "1", []Family{SourceText, Structure}},
		{Reference, "1", []Family{SymbolType}}, {Call, "1", []Family{SymbolType, Reference}},
		{DataFlow, "1", []Family{Reference, Call}},
		{SystemGraph, "1", []Family{Structure, DataFlow}},
		{HistoryChange, "1", nil},
		{Context, "1", []Family{SystemGraph, HistoryChange}},
		{Vector, "1", []Family{SystemGraph, Context}},
	}}
}
func (r Registry) Validate() error { _, err := r.TopologicalOrder(); return err }
func (r Registry) TopologicalOrder() ([]Family, error) {
	contracts := map[Family]Contract{}
	for _, c := range r.Contracts {
		if c.Family == "" || c.Version == "" {
			return nil, fmt.Errorf("projection contract requires family and version")
		}
		if _, ok := contracts[c.Family]; ok {
			return nil, fmt.Errorf("duplicate projection %q", c.Family)
		}
		contracts[c.Family] = c
	}
	for _, c := range r.Contracts {
		seen := map[Family]bool{}
		for _, dep := range c.Dependencies {
			if _, ok := contracts[dep]; !ok {
				return nil, fmt.Errorf("projection %q depends on unknown %q", c.Family, dep)
			}
			if seen[dep] {
				return nil, fmt.Errorf("duplicate dependency %q", dep)
			}
			seen[dep] = true
		}
	}
	names := make([]Family, 0, len(contracts))
	for f := range contracts {
		names = append(names, f)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	state := map[Family]int{}
	order := []Family{}
	var visit func(Family) error
	visit = func(f Family) error {
		if state[f] == 1 {
			return fmt.Errorf("projection dependency cycle at %q", f)
		}
		if state[f] == 2 {
			return nil
		}
		state[f] = 1
		deps := append([]Family{}, contracts[f].Dependencies...)
		sort.Slice(deps, func(i, j int) bool { return deps[i] < deps[j] })
		for _, d := range deps {
			if err := visit(d); err != nil {
				return err
			}
		}
		state[f] = 2
		order = append(order, f)
		return nil
	}
	for _, f := range names {
		if err := visit(f); err != nil {
			return nil, err
		}
	}
	return order, nil
}

// Downstream returns changed families and all transitive dependents in build
// order. Callers can mark these stale without rebuilding independent families.
func (r Registry) Downstream(changed ...Family) ([]Family, error) {
	order, err := r.TopologicalOrder()
	if err != nil {
		return nil, err
	}
	contracts := map[Family]Contract{}
	for _, c := range r.Contracts {
		contracts[c.Family] = c
	}
	affected := map[Family]bool{}
	for _, f := range changed {
		if _, ok := contracts[f]; !ok {
			return nil, fmt.Errorf("unknown projection %q", f)
		}
		affected[f] = true
	}
	result := []Family{}
	for _, f := range order {
		for _, dep := range contracts[f].Dependencies {
			if affected[dep] {
				affected[f] = true
			}
		}
		if affected[f] {
			result = append(result, f)
		}
	}
	return result, nil
}

// Fingerprints pin each family contract, family-specific profile, source input,
// and upstream outputs. Changes to one profile affect only its descendants.
func (r Registry) Fingerprints(source string, profiles map[Family]string) (map[Family]string, error) {
	order, err := r.TopologicalOrder()
	if err != nil {
		return nil, err
	}
	if source == "" {
		return nil, fmt.Errorf("source fingerprint required")
	}
	contracts := map[Family]Contract{}
	for _, c := range r.Contracts {
		contracts[c.Family] = c
	}
	for f := range profiles {
		if _, ok := contracts[f]; !ok {
			return nil, fmt.Errorf("profile for unknown projection %q", f)
		}
	}
	result := map[Family]string{}
	for _, f := range order {
		if profiles[f] == "" {
			return nil, fmt.Errorf("missing profile version for %q", f)
		}
		c := contracts[f]
		deps := map[Family]string{}
		for _, d := range c.Dependencies {
			deps[d] = result[d]
		}
		payload := struct {
			Algorithm                string
			Family                   Family
			Version, Source, Profile string
			Dependencies             map[Family]string
		}{"projection-input/v1", f, c.Version, source, profiles[f], deps}
		b, _ := json.Marshal(payload)
		h := sha256.Sum256(b)
		result[f] = hex.EncodeToString(h[:])
	}
	return result, nil
}

type Readiness struct {
	Known     int            `json:"known"`
	Requested int            `json:"requested"`
	Ready     int            `json:"ready"`
	Counts    map[Status]int `json:"counts"`
}

// Readiness counts every registered family in Known; omitted state is explicitly
// not_requested. Partial/unsupported/excluded results never count as ready.
func (r Registry) Readiness(states map[Family]Status) (Readiness, error) {
	if err := r.Validate(); err != nil {
		return Readiness{}, err
	}
	known := map[Family]bool{}
	for _, c := range r.Contracts {
		known[c.Family] = true
	}
	for f, s := range states {
		if !known[f] {
			return Readiness{}, fmt.Errorf("unknown projection %q", f)
		}
		if !s.Valid() {
			return Readiness{}, fmt.Errorf("invalid projection status %q", s)
		}
	}
	out := Readiness{Known: len(known), Counts: map[Status]int{}}
	for f := range known {
		s, ok := states[f]
		if !ok {
			s = NotRequested
		}
		out.Counts[s]++
		if s != NotRequested {
			out.Requested++
		}
		if s == Ready {
			out.Ready++
		}
	}
	return out, nil
}
