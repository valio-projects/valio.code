package snapshots

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
	"github.com/valio-projects/valio.code/internal/graph"
)

// buildGraphShards preserves cross-file edges while bounding each durable
// artifact by its source owner. Non-source nodes are placed with a deterministic
// file from their matching repository and project context; only the combined
// view is a complete graph value.
func buildGraphShards(ctx context.Context, v View, contents map[string]string) (map[string]json.RawMessage, error) {
	sources := []graph.Source{}
	for _, f := range v.Files {
		if f.Language == "go" {
			sources = append(sources, graph.Source{ID: f.ID, RepositoryID: string(f.RepositoryID), Path: f.Path, Content: contents[f.ID], ProjectIDs: f.ProjectIDs})
		}
	}
	output := map[string]json.RawMessage{}
	if len(sources) == 0 {
		return output, nil
	}
	g, err := graph.Build(ctx, sources)
	if err != nil {
		return nil, err
	}
	shards := map[string]*codegraph.Graph{}
	for _, s := range sources {
		shards[s.ID] = &codegraph.Graph{Schema: g.Schema, Nodes: []codegraph.Node{}, Edges: []codegraph.Edge{}, Diagnostics: []codegraph.Diagnostic{}}
	}
	first := sources[0].ID
	owner := map[string]string{}
	for _, n := range g.Nodes {
		id := n.FileID
		if id == "" {
			var ok bool
			id, ok = compatibleGraphShard(sources, n)
			if !ok {
				return nil, fmt.Errorf("code graph node %q has no compatible source shard", n.ID)
			}
		}
		if shards[id] == nil {
			return nil, fmt.Errorf("code graph node %q names unknown source shard %q", n.ID, id)
		}
		owner[n.ID] = id
		shards[id].Nodes = append(shards[id].Nodes, n)
	}
	for _, e := range g.Edges {
		id := owner[e.SourceID]
		if id == "" || shards[id] == nil {
			return nil, fmt.Errorf("code graph edge %q has no owned source node", e.ID)
		}
		shards[id].Edges = append(shards[id].Edges, e)
	}
	for _, d := range g.Diagnostics {
		id := first
		if d.Range != nil {
			id = d.Range.FileID
		}
		shards[id].Diagnostics = append(shards[id].Diagnostics, d)
	}
	shards[first].Completeness = g.Completeness
	for id, shard := range shards {
		b, e := json.Marshal(shard)
		if e != nil {
			return nil, e
		}
		output[id] = b
	}
	return output, nil
}

// compatibleGraphShard selects the lowest stable source ID in the same
// repository whose project contexts include the non-source node's contexts.
// It deliberately does not invent a project association for a package variant.
func compatibleGraphShard(sources []graph.Source, node codegraph.Node) (string, bool) {
	ids := make([]string, 0, len(sources))
	for _, source := range sources {
		if source.RepositoryID != node.RepositoryID || !includesProjects(source.ProjectIDs, node.ProjectIDs) {
			continue
		}
		ids = append(ids, source.ID)
	}
	if len(ids) == 0 {
		return "", false
	}
	sort.Strings(ids)
	return ids[0], true
}

func includesProjects(values, required []string) bool {
	available := make(map[string]bool, len(values))
	for _, value := range values {
		available[value] = true
	}
	for _, value := range required {
		if !available[value] {
			return false
		}
	}
	return true
}
