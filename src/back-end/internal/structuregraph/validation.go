package structuregraph

import (
	"fmt"
)

// Validate checks identity, scope, range, and relation invariants without
// interpreting a written name as a semantic target.
func (g Graph) Validate() error {
	if g.Schema != SchemaVersion || g.Language == "" {
		return fmt.Errorf("unsupported structure graph schema or language")
	}
	nodes := map[string]Node{}
	for _, node := range g.Nodes {
		if node.ID == "" || !validNodeKind(node.Kind) || node.FileID == "" || node.RepositoryID == "" {
			return fmt.Errorf("invalid structure graph node %q", node.ID)
		}
		if node.Range.FileID != node.FileID || node.Range.Start < 0 || node.Range.End < node.Range.Start {
			return fmt.Errorf("invalid structure graph node range %q", node.ID)
		}
		if node.Evidence.Producer == "" || node.Evidence.Range == nil || !sameRange(*node.Evidence.Range, node.Range) {
			return fmt.Errorf("invalid structure graph node evidence %q", node.ID)
		}
		if _, exists := nodes[node.ID]; exists {
			return fmt.Errorf("duplicate structure graph node %q", node.ID)
		}
		nodes[node.ID] = node
	}
	edges := map[string]bool{}
	for _, edge := range g.Edges {
		if edge.ID == "" || !validRelationKind(edge.Kind) || edge.SourceID == "" || nodes[edge.SourceID].ID == "" || edges[edge.ID] {
			return fmt.Errorf("invalid structure graph edge %q", edge.ID)
		}
		if edge.Evidence.Producer == "" || edge.Evidence.Range == nil || edge.Evidence.Range.FileID != nodes[edge.SourceID].FileID || edge.Evidence.Range.Start < 0 || edge.Evidence.Range.End < edge.Evidence.Range.Start {
			return fmt.Errorf("invalid structure graph edge evidence %q", edge.ID)
		}
		if !sameStrings(edge.ProjectIDs, nodes[edge.SourceID].ProjectIDs) {
			return fmt.Errorf("invalid structure graph edge scope %q", edge.ID)
		}
		switch edge.Resolution {
		case ResolutionExact:
			target := nodes[edge.TargetID]
			if edge.Kind == RelationImports || edge.TargetID == "" || target.ID == "" || target.FileID != nodes[edge.SourceID].FileID || target.RepositoryID != nodes[edge.SourceID].RepositoryID || !sameStrings(target.ProjectIDs, nodes[edge.SourceID].ProjectIDs) {
				return fmt.Errorf("exact structure graph edge %q needs a local target", edge.ID)
			}
		case ResolutionUnresolved:
			if edge.TargetID != "" || edge.Kind != RelationImports {
				return fmt.Errorf("invalid unresolved structure graph edge %q", edge.ID)
			}
		default:
			return fmt.Errorf("invalid structure graph edge resolution %q", edge.ID)
		}
		edges[edge.ID] = true
	}
	return nil
}

func validNodeKind(kind NodeKind) bool {
	switch kind {
	case NodeFile, NodeDeclaration, NodeParameter, NodeImport, NodeReference, NodeCall:
		return true
	default:
		return false
	}
}

func validRelationKind(kind RelationKind) bool {
	switch kind {
	case RelationContains, RelationDeclares, RelationHasParameter, RelationImports:
		return true
	default:
		return false
	}
}

func sameRange(left, right ByteRange) bool {
	return left.FileID == right.FileID && left.Start == right.Start && left.End == right.End
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
