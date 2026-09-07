package codegraph

import "fmt"

// Validate verifies graph-local identity, range and relation invariants.
func (g Graph) Validate() error {
	if g.Schema != SchemaVersion {
		return fmt.Errorf("unsupported graph schema %q", g.Schema)
	}
	nodes := map[string]bool{}
	for _, node := range g.Nodes {
		if node.ID == "" || node.Kind == "" || node.RepositoryID == "" || nodes[node.ID] {
			return fmt.Errorf("invalid or duplicate node %q", node.ID)
		}
		if node.Range != nil && (node.Range.FileID == "" || node.Range.Start < 0 || node.Range.End < node.Range.Start) {
			return fmt.Errorf("invalid node range %q", node.ID)
		}
		nodes[node.ID] = true
	}
	edges := map[string]bool{}
	for _, edge := range g.Edges {
		if edge.ID == "" || edge.Kind == "" || edge.SourceID == "" || !nodes[edge.SourceID] || edges[edge.ID] {
			return fmt.Errorf("invalid or duplicate edge %q", edge.ID)
		}
		if edge.Resolution == ResolutionExact && (edge.TargetID == "" || !nodes[edge.TargetID]) {
			return fmt.Errorf("exact edge %q requires a local target", edge.ID)
		}
		if edge.Resolution == ResolutionUnresolved && edge.TargetID != "" {
			return fmt.Errorf("unresolved edge %q cannot have a target", edge.ID)
		}
		if edge.Resolution != ResolutionExact && edge.Resolution != ResolutionCandidate && edge.Resolution != ResolutionUnresolved {
			return fmt.Errorf("invalid edge resolution %q", edge.Resolution)
		}
		edges[edge.ID] = true
	}
	return nil
}
