package projects

import (
	"encoding/json"
	"fmt"
	"github.com/valio-projects/valio.code/internal/domain"
)

// MembershipResolver owns a copied set of definitions so caller edits cannot
// change membership between reads.
type MembershipResolver struct{ definitions []Definition }

// NewMembershipResolver validates and copies definitions for stable later reads.
func NewMembershipResolver(definitions []Definition) (*MembershipResolver, error) {
	if _, err := resolveMembership(definitions, nil); err != nil {
		return nil, err
	}
	data, err := json.Marshal(definitions)
	if err != nil {
		return nil, err
	}
	var copied []Definition
	if err := json.Unmarshal(data, &copied); err != nil {
		return nil, err
	}
	return &MembershipResolver{definitions: copied}, nil
}

// Resolve returns sorted project memberships for each file; files outside roots have empty slices.
func (r *MembershipResolver) Resolve(files []FileRef) (map[FileRef][]domain.ProjectID, error) {
	if r == nil {
		return nil, fmt.Errorf("membership resolver is required")
	}
	return resolveMembership(r.definitions, files)
}
