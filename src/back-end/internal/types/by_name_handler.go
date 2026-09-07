package types

import (
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
	"sort"
)

// ByNameHandler serves a pure read query against a validated immutable catalog.
type ByNameHandler struct{ catalog *Catalog }

// NewByNameHandler creates a handler over catalog; a nil catalog is rejected by Handle.
func NewByNameHandler(catalog *Catalog) *ByNameHandler { return &ByNameHandler{catalog: catalog} }

// Handle performs exact, case-sensitive matching and preserves ambiguity.
func (h *ByNameHandler) Handle(query ByNameQuery) (NameResult, error) {
	if h == nil || h.catalog == nil {
		return NameResult{}, fmt.Errorf("name query handler requires a catalog")
	}
	c, scope, name := h.catalog, query.Scope, query.Name
	if scope.WorkspaceID == "" || name == "" {
		return NameResult{}, fmt.Errorf("name lookup requires workspace and name")
	}
	selected := []typeinfo.TypeDescriptor{}
	for _, d := range c.descriptors {
		if d.Scope.WorkspaceID != scope.WorkspaceID || (scope.ProjectID != "" && d.Scope.ProjectID != scope.ProjectID) || (scope.BuildProfileID != "" && d.Scope.BuildProfileID != scope.BuildProfileID) || (scope.VersionID != "" && d.Scope.VersionID != scope.VersionID) {
			continue
		}
		simple := d.Name.Value != nil && *d.Name.Value == name
		full := d.FullyQualifiedName.EffectiveStatus() == typeinfo.FactKnown && d.FullyQualifiedName.Value != nil && *d.FullyQualifiedName.Value == name
		if simple || full {
			selected = append(selected, d)
		}
	}
	sort.Slice(selected, func(i, j int) bool {
		a, b := selected[i], selected[j]
		if a.Scope.ProjectID != b.Scope.ProjectID {
			return a.Scope.ProjectID < b.Scope.ProjectID
		}
		if a.Scope.BuildProfileID != b.Scope.BuildProfileID {
			return a.Scope.BuildProfileID < b.Scope.BuildProfileID
		}
		if a.Scope.VersionID != b.Scope.VersionID {
			return a.Scope.VersionID < b.Scope.VersionID
		}
		return a.ID < b.ID
	})
	selected, err := clone(selected)
	if err != nil {
		return NameResult{}, err
	}
	result := NameResult{Status: MatchNotFound, Candidates: []Candidate{}, Reference: typeinfo.SymbolReference{Resolution: typeinfo.ReferenceUnresolved, Candidates: []typeinfo.SymbolIdentity{}, Reason: "no type matches the requested scope and name"}}
	refs := []typeinfo.SymbolIdentity{}
	for _, d := range selected {
		id := typeinfo.SymbolIdentity{ID: d.ID, Scope: d.Scope}
		refs = append(refs, id)
		result.Candidates = append(result.Candidates, Candidate{Type: d, Counts: Counts(d), Reference: typeinfo.SymbolReference{Resolution: typeinfo.ReferenceExact, Exact: &id, Candidates: []typeinfo.SymbolIdentity{}}})
	}
	if len(refs) == 1 {
		result.Status = MatchExact
		result.Reference = typeinfo.SymbolReference{Resolution: typeinfo.ReferenceExact, Exact: &refs[0], Candidates: []typeinfo.SymbolIdentity{}}
	} else if len(refs) > 1 {
		result.Status = MatchAmbiguous
		result.Reference = typeinfo.SymbolReference{Resolution: typeinfo.ReferenceCandidate, Candidates: refs, Reason: "multiple versioned types match; narrow project, build profile, version or qualified name"}
	}
	return result, nil
}
