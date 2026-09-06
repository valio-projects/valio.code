package types

import (
	"encoding/json"
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

type Catalog struct{ descriptors []typeinfo.TypeDescriptor }

// ByName is a convenience entry point to the explicit query handler.
func (c *Catalog) ByName(scope QueryScope, name string) (NameResult, error) {
	return NewByNameHandler(c).Handle(ByNameQuery{Scope: scope, Name: name})
}

func NewCatalog(descriptors []typeinfo.TypeDescriptor) (*Catalog, error) {
	seen := map[typeinfo.SymbolIdentity]bool{}
	for _, d := range descriptors {
		if err := d.Validate(); err != nil {
			return nil, fmt.Errorf("type %q: %w", d.ID, err)
		}
		key := typeinfo.SymbolIdentity{ID: d.ID, Scope: d.Scope}
		if seen[key] {
			return nil, fmt.Errorf("duplicate versioned type identity %q", d.ID)
		}
		seen[key] = true
	}
	cloned, err := clone(descriptors)
	if err != nil {
		return nil, err
	}
	return &Catalog{descriptors: cloned}, nil
}
func clone(descriptors []typeinfo.TypeDescriptor) ([]typeinfo.TypeDescriptor, error) {
	b, err := json.Marshal(descriptors)
	if err != nil {
		return nil, err
	}
	var result []typeinfo.TypeDescriptor
	if err := json.Unmarshal(b, &result); err != nil {
		return nil, err
	}
	return result, nil
}
