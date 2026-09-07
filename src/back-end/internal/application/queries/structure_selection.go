package queries

import (
	"context"
	"github.com/valio-projects/valio.code/internal/structuregraph"
)

// selectStructure follows syntax ownership only; it does not resolve call targets.
func selectStructure(ctx context.Context, g structuregraph.Graph, q StructureQuery, nodeLimit, edgeLimit int) (StructureFileResult, []string) {
	result := StructureFileResult{Nodes: []structuregraph.Node{}, Edges: []structuregraph.Edge{}}
	reasons := []string{}
	selected := map[string]bool{}
	frontier := []string{}
	for _, node := range g.Nodes {
		match := q.NodeID != "" && node.ID == q.NodeID || q.Name != "" && node.Kind == structuregraph.NodeDeclaration && node.Name == q.Name || q.NodeID == "" && q.Name == "" && node.Kind == structuregraph.NodeFile
		if match {
			if len(selected) >= nodeLimit {
				reasons = append(reasons, "NODE_BUDGET")
				break
			}
			selected[node.ID] = true
			frontier = append(frontier, node.ID)
		}
	}
	outgoing := map[string][]string{}
	for _, edge := range g.Edges {
		if edge.TargetID != "" {
			outgoing[edge.SourceID] = append(outgoing[edge.SourceID], edge.TargetID)
		}
	}
	for depth := 0; depth < q.Depth && len(frontier) > 0; depth++ {
		next := []string{}
		for _, id := range frontier {
			if ctx.Err() != nil {
				return result, reasons
			}
			for _, target := range outgoing[id] {
				if !selected[target] {
					if len(selected) >= nodeLimit {
						reasons = append(reasons, "NODE_BUDGET")
						break
					}
					selected[target] = true
					next = append(next, target)
				}
			}
		}
		frontier = next
	}
	for _, id := range frontier {
		for _, target := range outgoing[id] {
			if !selected[target] {
				reasons = append(reasons, "DEPTH_BUDGET")
				break
			}
		}
		if len(reasons) > 0 {
			break
		}
	}
	for _, node := range g.Nodes {
		if selected[node.ID] {
			result.Nodes = append(result.Nodes, node)
		}
	}
	for _, edge := range g.Edges {
		if selected[edge.SourceID] && (edge.TargetID == "" || selected[edge.TargetID]) {
			if len(result.Edges) >= edgeLimit {
				reasons = append(reasons, "EDGE_BUDGET")
				break
			}
			result.Edges = append(result.Edges, edge)
		}
	}
	return result, reasons
}
