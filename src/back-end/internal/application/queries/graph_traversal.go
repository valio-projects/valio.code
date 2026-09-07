package queries

import (
	"sort"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
)

type scopedGraph struct {
	nodes       map[string]codegraph.Node
	edges       []codegraph.Edge
	diagnostics []codegraph.Diagnostic
}

func scopeGraph(graph codegraph.Graph, projects []string) scopedGraph {
	selected := map[string]bool{}
	for _, projectID := range projects {
		selected[projectID] = true
	}
	nodes := map[string]codegraph.Node{}
	for _, node := range graph.Nodes {
		if len(selected) == 0 || intersects(node.ProjectIDs, selected) {
			nodes[node.ID] = node
		}
	}
	edges := make([]codegraph.Edge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		if _, ok := nodes[edge.SourceID]; !ok {
			continue
		}
		if edge.TargetID != "" {
			if _, ok := nodes[edge.TargetID]; !ok {
				continue
			}
		}
		if len(selected) == 0 || intersects(edge.ProjectIDs, selected) {
			edges = append(edges, edge)
		}
	}
	diagnostics := []codegraph.Diagnostic{}
	for _, diagnostic := range graph.Diagnostics {
		if diagnostic.Range == nil {
			if len(selected) == 0 {
				diagnostics = append(diagnostics, diagnostic)
			}
			continue
		}
		if len(selected) == 0 {
			diagnostics = append(diagnostics, diagnostic)
			continue
		}
		for _, node := range nodes {
			if node.FileID == diagnostic.Range.FileID {
				diagnostics = append(diagnostics, diagnostic)
				break
			}
		}
	}
	return scopedGraph{nodes: nodes, edges: edges, diagnostics: diagnostics}
}

func intersects(values []string, selected map[string]bool) bool {
	for _, value := range values {
		if selected[value] {
			return true
		}
	}
	return false
}

func executeGraphQuery(query GraphQuery, graph scopedGraph) GraphResult {
	result := GraphResult{Diagnostics: graph.diagnostics}
	limit := query.Limit
	if limit == 0 {
		limit = DefaultGraphTargets
	}
	depth := query.Depth
	if depth == 0 {
		depth = DefaultGraphDepth
	}
	switch query.Mode {
	case GraphSymbols:
		for _, node := range sortedScopedNodes(graph.nodes) {
			if node.Kind == codegraph.NodeSymbol && node.Name == query.Name {
				if !addResultNode(&result, node) {
					break
				}
				if len(result.Nodes) == limit {
					if hasMoreSymbols(graph.nodes, query.Name, len(result.Nodes)) {
						addBudgetNotice(&result)
					}
					break
				}
			}
		}
	case GraphReferences:
		directIncoming(&result, graph, query.TargetNodeID, codegraph.RelationRefersTo, query.EdgeKinds, limit)
	case GraphCallers:
		directIncoming(&result, graph, query.TargetNodeID, codegraph.RelationCalls, query.EdgeKinds, limit)
	case GraphCallees:
		directOutgoing(&result, graph, query.TargetNodeID, codegraph.RelationCalls, query.EdgeKinds, limit)
	case GraphReads:
		directIncoming(&result, graph, query.TargetNodeID, codegraph.RelationReads, query.EdgeKinds, limit)
	case GraphWrites:
		directIncoming(&result, graph, query.TargetNodeID, codegraph.RelationWrites, query.EdgeKinds, limit)
	case GraphNeighbors:
		traverseNeighbors(&result, graph, query.TargetNodeID, query.EdgeKinds, depth)
	case GraphPaths:
		findPaths(&result, graph, query.SourceNodeID, query.TargetNodeID, query.EdgeKinds, depth)
	}
	return result
}

func directIncoming(result *GraphResult, graph scopedGraph, targetID string, required codegraph.RelationKind, selected []codegraph.RelationKind, limit int) {
	target, ok := graph.nodes[targetID]
	if !ok {
		return
	}
	addResultNode(result, target)
	matches := []codegraph.Edge{}
	for _, edge := range graph.edges {
		if edge.TargetID != targetID || edge.Kind != required || !edgeKindSelected(edge.Kind, selected) {
			continue
		}
		matches = append(matches, edge)
	}
	for index, edge := range matches {
		if index == limit {
			addBudgetNotice(result)
			return
		}
		if !addResultEdge(result, edge) || !addResultNode(result, graph.nodes[edge.SourceID]) {
			return
		}
	}
}

func directOutgoing(result *GraphResult, graph scopedGraph, sourceID string, required codegraph.RelationKind, selected []codegraph.RelationKind, limit int) {
	source, ok := graph.nodes[sourceID]
	if !ok {
		return
	}
	addResultNode(result, source)
	if source.Kind == codegraph.NodeSymbol && required == codegraph.RelationCalls {
		directCallableCallees(result, graph, sourceID, selected, limit)
		return
	}
	matches := []codegraph.Edge{}
	for _, edge := range graph.edges {
		if edge.SourceID != sourceID || edge.TargetID == "" || edge.Kind != required || !edgeKindSelected(edge.Kind, selected) {
			continue
		}
		matches = append(matches, edge)
	}
	for index, edge := range matches {
		if index == limit {
			addBudgetNotice(result)
			return
		}
		if !addResultEdge(result, edge) || !addResultNode(result, graph.nodes[edge.TargetID]) {
			return
		}
	}
}

func directCallableCallees(result *GraphResult, graph scopedGraph, symbolID string, selected []codegraph.RelationKind, limit int) {
	count := 0
	for _, containment := range graph.edges {
		if containment.Kind != codegraph.RelationContains || containment.SourceID != symbolID || graph.nodes[containment.TargetID].Kind != codegraph.NodeCallSite {
			continue
		}
		for _, call := range graph.edges {
			if call.Kind != codegraph.RelationCalls || call.SourceID != containment.TargetID || call.TargetID == "" || !edgeKindSelected(call.Kind, selected) {
				continue
			}
			if count == limit {
				addBudgetNotice(result)
				return
			}
			if !addResultEdge(result, containment) || !addResultNode(result, graph.nodes[containment.TargetID]) || !addResultEdge(result, call) || !addResultNode(result, graph.nodes[call.TargetID]) {
				return
			}
			count++
		}
	}
}

func traverseNeighbors(result *GraphResult, graph scopedGraph, startID string, selected []codegraph.RelationKind, depth int) {
	start, ok := graph.nodes[startID]
	if !ok {
		return
	}
	if !addResultNode(result, start) {
		return
	}
	type step struct {
		id    string
		depth int
	}
	queue := []step{{id: startID}}
	seen := map[string]bool{startID: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.depth == depth {
			continue
		}
		for _, edge := range graph.edges {
			if !edgeKindSelected(edge.Kind, selected) || edge.TargetID == "" {
				continue
			}
			next := ""
			if edge.SourceID == current.id {
				next = edge.TargetID
			} else if edge.TargetID == current.id {
				next = edge.SourceID
			}
			if next == "" {
				continue
			}
			if !addResultEdge(result, edge) || !addResultNode(result, graph.nodes[next]) {
				return
			}
			if !seen[next] {
				seen[next] = true
				queue = append(queue, step{id: next, depth: current.depth + 1})
			}
		}
	}
}

func findPaths(result *GraphResult, graph scopedGraph, sourceID, targetID string, selected []codegraph.RelationKind, depth int) {
	if graph.nodes[sourceID].ID == "" || graph.nodes[targetID].ID == "" {
		return
	}
	type state struct {
		nodes []string
		edges []codegraph.Edge
	}
	queue := []state{{nodes: []string{sourceID}}}
	for len(queue) > 0 && len(result.Paths) < MaxGraphPaths {
		current := queue[0]
		queue = queue[1:]
		if len(current.edges) == depth {
			continue
		}
		last := current.nodes[len(current.nodes)-1]
		for _, edge := range graph.edges {
			if edge.SourceID != last || edge.TargetID == "" || !edgeKindSelected(edge.Kind, selected) || containsPathNode(current.nodes, edge.TargetID) {
				continue
			}
			nextNodes := append(append([]string{}, current.nodes...), edge.TargetID)
			nextEdges := append(append([]codegraph.Edge{}, current.edges...), edge)
			if edge.TargetID == targetID {
				if !addPath(result, graph, nextNodes, nextEdges) {
					return
				}
				continue
			}
			queue = append(queue, state{nodes: nextNodes, edges: nextEdges})
		}
	}
	if len(queue) > 0 {
		addBudgetNotice(result)
	}
}

func containsPathNode(nodes []string, id string) bool {
	for _, node := range nodes {
		if node == id {
			return true
		}
	}
	return false
}

func addPath(result *GraphResult, graph scopedGraph, nodes []string, edges []codegraph.Edge) bool {
	if len(result.Paths) >= MaxGraphPaths {
		addBudgetNotice(result)
		return false
	}
	for _, nodeID := range nodes {
		if !addResultNode(result, graph.nodes[nodeID]) {
			return false
		}
	}
	for _, edge := range edges {
		if !addResultEdge(result, edge) {
			return false
		}
	}
	result.Paths = append(result.Paths, nodes)
	return true
}

func addResultNode(result *GraphResult, node codegraph.Node) bool {
	for _, existing := range result.Nodes {
		if existing.ID == node.ID {
			return true
		}
	}
	if len(result.Nodes) >= MaxGraphNodes {
		addBudgetNotice(result)
		return false
	}
	result.Nodes = append(result.Nodes, node)
	return true
}

func addResultEdge(result *GraphResult, edge codegraph.Edge) bool {
	for _, existing := range result.Edges {
		if existing.ID == edge.ID {
			return true
		}
	}
	if len(result.Edges) >= MaxGraphEdges {
		addBudgetNotice(result)
		return false
	}
	result.Edges = append(result.Edges, edge)
	if edge.Resolution == codegraph.ResolutionCandidate {
		addNotice(result, GraphNotice{Code: GraphCandidateTarget, Message: "a returned relation has a candidate target rather than an exact resolution"})
	}
	return true
}

func addBudgetNotice(result *GraphResult) {
	addNotice(result, GraphNotice{Code: GraphNodeBudget, Message: "the bounded graph response reached a result limit"})
	result.Status = GraphPartial
}

func addNotice(result *GraphResult, notice GraphNotice) {
	for _, existing := range result.Notices {
		if existing.Code == notice.Code {
			return
		}
	}
	result.Notices = append(result.Notices, notice)
}

func edgeKindSelected(kind codegraph.RelationKind, selected []codegraph.RelationKind) bool {
	if len(selected) == 0 {
		return true
	}
	for _, candidate := range selected {
		if candidate == kind {
			return true
		}
	}
	return false
}

func hasMoreSymbols(nodes map[string]codegraph.Node, name string, found int) bool {
	count := 0
	for _, node := range nodes {
		if node.Kind == codegraph.NodeSymbol && node.Name == name {
			count++
		}
	}
	return count > found
}

func sortedScopedNodes(nodes map[string]codegraph.Node) []codegraph.Node {
	result := make([]codegraph.Node, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, node)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func sortGraphResult(result *GraphResult) {
	sort.Slice(result.Nodes, func(i, j int) bool { return result.Nodes[i].ID < result.Nodes[j].ID })
	sort.Slice(result.Edges, func(i, j int) bool { return result.Edges[i].ID < result.Edges[j].ID })
	sort.Slice(result.Paths, func(i, j int) bool {
		for index := 0; index < len(result.Paths[i]) && index < len(result.Paths[j]); index++ {
			if result.Paths[i][index] != result.Paths[j][index] {
				return result.Paths[i][index] < result.Paths[j][index]
			}
		}
		return len(result.Paths[i]) < len(result.Paths[j])
	})
}
