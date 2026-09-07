package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain/codegraph"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
)

// graphArtifactStore is the read-only subset required after Service resolves a
// view. Service.Store remains the broader publication repository at its API boundary.
type graphArtifactStore interface {
	Artifacts(context.Context, snapshots.View) ([]snapshots.Artifact, error)
}

type mergedGraph struct {
	graph       codegraph.Graph
	missing     []snapshots.FileRef
	unsupported bool
}

// Graph resolves a view once, merges its per-file graph shards, verifies every
// graph fact against pinned file and project membership, then performs one
// bounded graph operation. It never reads source content or recomputes graph facts.
func (s Service) Graph(ctx context.Context, q GraphQuery) (GraphResult, error) {
	result := GraphResult{}
	if err := validateGraphQuery(s, q); err != nil {
		return result, err
	}
	view, err := s.View(ctx, q.Scope.ViewID)
	if err != nil {
		return result, err
	}
	if !viewHasProjects(view, q.Scope.ProjectIDs) {
		return result, fault.ErrNotFound
	}
	store, ok := s.Store.(graphArtifactStore)
	if !ok {
		return result, fault.ErrInvalid
	}
	artifacts, err := store.Artifacts(ctx, view)
	if err != nil {
		return result, err
	}
	merged, err := mergeGraphArtifacts(view, artifacts)
	if err != nil {
		return result, err
	}
	if merged.unsupported {
		return GraphResult{ViewID: view.ID, Status: GraphUnsupported, Notices: []GraphNotice{{Code: GraphUnsupportedOldView, Message: "the selected view has no compatible code graph projection"}}}, nil
	}
	scoped := scopeGraph(merged.graph, q.Scope.ProjectIDs)
	result = executeGraphQuery(q, scoped)
	result.ViewID = view.ID
	result.Completeness = merged.graph.Completeness
	if missingGraphInScope(merged.missing, q.Scope.ProjectIDs) || graphIsPartial(merged.graph.Completeness) || result.Status == GraphPartial {
		result.Status = GraphPartial
	} else {
		result.Status = GraphComplete
	}
	if missingGraphInScope(merged.missing, q.Scope.ProjectIDs) {
		result.Notices = append(result.Notices, GraphNotice{Code: GraphMissingShard, Message: "one or more selected Go files have no graph shard"})
	}
	sortGraphResult(&result)
	return result, nil
}

func validateGraphQuery(s Service, q GraphQuery) error {
	if q.Scope.WorkspaceID != s.WorkspaceID {
		return fault.ErrForbidden
	}
	if len(q.Scope.ProjectIDs) > 128 || q.Name != strings.TrimSpace(q.Name) || len(q.Name) > 4096 || q.Limit < 0 || q.Limit > MaxGraphTargets || q.Depth < 0 || q.Depth > MaxGraphDepth {
		return fault.ErrInvalid
	}
	seenProjects := map[string]bool{}
	for _, projectID := range q.Scope.ProjectIDs {
		if projectID == "" || seenProjects[projectID] {
			return fault.ErrInvalid
		}
		seenProjects[projectID] = true
	}
	seenKinds := map[codegraph.RelationKind]bool{}
	for _, kind := range q.EdgeKinds {
		if !validRelationKind(kind) || seenKinds[kind] {
			return fault.ErrInvalid
		}
		seenKinds[kind] = true
	}
	switch q.Mode {
	case GraphSymbols:
		if q.Name == "" || q.TargetNodeID != "" || q.SourceNodeID != "" {
			return fault.ErrInvalid
		}
	case GraphReferences, GraphCallers, GraphCallees, GraphReads, GraphWrites, GraphNeighbors:
		if q.TargetNodeID == "" || q.SourceNodeID != "" {
			return fault.ErrInvalid
		}
	case GraphPaths:
		if q.SourceNodeID == "" || q.TargetNodeID == "" || q.SourceNodeID == q.TargetNodeID {
			return fault.ErrInvalid
		}
	default:
		return fault.ErrInvalid
	}
	return nil
}

func validRelationKind(kind codegraph.RelationKind) bool {
	switch kind {
	case codegraph.RelationContains, codegraph.RelationDeclares, codegraph.RelationRefersTo, codegraph.RelationImports, codegraph.RelationMemberOf, codegraph.RelationHasType, codegraph.RelationCalls, codegraph.RelationReads, codegraph.RelationWrites:
		return true
	default:
		return false
	}
}

func viewHasProjects(view snapshots.View, requested []string) bool {
	available := map[string]bool{}
	for _, definition := range view.Projects {
		available[string(definition.Project.ID)] = true
	}
	for _, projectID := range requested {
		if !available[projectID] {
			return false
		}
	}
	return true
}

func mergeGraphArtifacts(view snapshots.View, artifacts []snapshots.Artifact) (mergedGraph, error) {
	files := map[string]snapshots.FileRef{}
	repositories := map[string]bool{}
	for _, file := range view.Files {
		files[file.ID] = file
		repositories[string(file.RepositoryID)] = true
	}
	projects := map[string]bool{}
	for _, definition := range view.Projects {
		projects[string(definition.Project.ID)] = true
	}
	merged := codegraph.Graph{Schema: codegraph.SchemaVersion}
	nodes := map[string]codegraph.Node{}
	edges := map[string]codegraph.Edge{}
	edgeOwners := map[string]string{}
	diagnostics := map[string]codegraph.Diagnostic{}
	nodeOwners := map[string]string{}
	artifactFiles := map[string]bool{}
	missing := map[string]snapshots.FileRef{}
	found := false
	for _, artifact := range artifacts {
		if artifact.ViewID != view.ID {
			return mergedGraph{}, fault.ErrForbidden
		}
		file, ok := files[artifact.FileID]
		if !ok || artifactFiles[artifact.FileID] {
			return mergedGraph{}, fault.ErrForbidden
		}
		artifactFiles[artifact.FileID] = true
		if len(artifact.Graph) == 0 {
			if file.Language == "go" {
				missing[file.ID] = file
			}
			continue
		}
		var shard codegraph.Graph
		if err := json.Unmarshal(artifact.Graph, &shard); err != nil {
			return mergedGraph{}, fault.ErrInvalid
		}
		if shard.Schema != codegraph.SchemaVersion {
			return mergedGraph{unsupported: true}, nil
		}
		found = true
		if err := verifyShard(shard, artifact.FileID, file, files, projects, repositories); err != nil {
			return mergedGraph{}, err
		}
		for _, node := range shard.Nodes {
			if prior, exists := nodes[node.ID]; exists {
				if !reflect.DeepEqual(prior, node) || nodeOwners[node.ID] != artifact.FileID {
					return mergedGraph{}, fault.ErrInvalid
				}
				continue
			}
			nodes[node.ID] = node
			nodeOwners[node.ID] = artifact.FileID
		}
		for _, edge := range shard.Edges {
			if prior, exists := edges[edge.ID]; exists {
				if !reflect.DeepEqual(prior, edge) || edgeOwners[edge.ID] != artifact.FileID {
					return mergedGraph{}, fault.ErrInvalid
				}
				continue
			}
			edges[edge.ID] = edge
			edgeOwners[edge.ID] = artifact.FileID
		}
		for _, diagnostic := range shard.Diagnostics {
			diagnostics[diagnosticKey(diagnostic)] = diagnostic
		}
		if completenessPresent(shard.Completeness) {
			merged.Completeness = shard.Completeness
		}
	}
	for _, file := range view.Files {
		if file.Language == "go" && !artifactFiles[file.ID] {
			missing[file.ID] = file
		}
	}
	if !found {
		return mergedGraph{unsupported: true}, nil
	}
	for _, node := range nodes {
		if node.FileID == "" {
			continue
		}
		if nodeOwners[node.ID] != node.FileID {
			return mergedGraph{}, fault.ErrForbidden
		}
	}
	for _, edge := range edges {
		source, ok := nodes[edge.SourceID]
		if !ok || nodeOwners[edge.SourceID] != edgeOwners[edge.ID] {
			return mergedGraph{}, fault.ErrInvalid
		}
		if edge.TargetID != "" {
			target, ok := nodes[edge.TargetID]
			if !ok || !subset(edge.ProjectIDs, source.ProjectIDs) || !subset(edge.ProjectIDs, target.ProjectIDs) {
				return mergedGraph{}, fault.ErrForbidden
			}
		} else if !subset(edge.ProjectIDs, source.ProjectIDs) {
			return mergedGraph{}, fault.ErrForbidden
		}
	}
	merged.Nodes = mapNodes(nodes)
	merged.Edges = mapEdges(edges)
	merged.Diagnostics = mapDiagnostics(diagnostics)
	if err := merged.Validate(); err != nil {
		return mergedGraph{}, fault.ErrInvalid
	}
	missingFiles := make([]snapshots.FileRef, 0, len(missing))
	for _, file := range missing {
		missingFiles = append(missingFiles, file)
	}
	return mergedGraph{graph: merged, missing: missingFiles}, nil
}

func missingGraphInScope(files []snapshots.FileRef, projects []string) bool {
	if len(files) == 0 {
		return false
	}
	if len(projects) == 0 {
		return true
	}
	selected := map[string]bool{}
	for _, projectID := range projects {
		selected[projectID] = true
	}
	for _, file := range files {
		if intersects(file.ProjectIDs, selected) {
			return true
		}
	}
	return false
}

func verifyShard(shard codegraph.Graph, artifactFileID string, artifactFile snapshots.FileRef, files map[string]snapshots.FileRef, projects map[string]bool, repositories map[string]bool) error {
	for _, node := range shard.Nodes {
		if node.ID == "" || node.Kind == "" || node.RepositoryID == "" || !repositories[node.RepositoryID] || node.RepositoryID != string(artifactFile.RepositoryID) || !subsetOfKnown(node.ProjectIDs, projects) || !subset(node.ProjectIDs, artifactFile.ProjectIDs) {
			return fault.ErrForbidden
		}
		if node.FileID != "" && (node.FileID != artifactFileID || files[node.FileID].ID == "") {
			return fault.ErrForbidden
		}
		if node.Range != nil && (node.Range.FileID != node.FileID || node.Range.Start < 0 || node.Range.End < node.Range.Start) {
			return fault.ErrInvalid
		}
		if node.Evidence.Range != nil && (node.Evidence.Range.FileID != node.FileID || node.Evidence.Range.Start < 0 || node.Evidence.Range.End < node.Evidence.Range.Start) {
			return fault.ErrInvalid
		}
	}
	for _, edge := range shard.Edges {
		if !validRelationKind(edge.Kind) {
			return fault.ErrInvalid
		}
		if edge.Evidence.Range != nil && (edge.Evidence.Range.FileID != artifactFileID || edge.Evidence.Range.Start < 0 || edge.Evidence.Range.End < edge.Evidence.Range.Start) {
			return fault.ErrForbidden
		}
	}
	for _, diagnostic := range shard.Diagnostics {
		if diagnostic.Range != nil && diagnostic.Range.FileID != artifactFileID {
			return fault.ErrForbidden
		}
	}
	return nil
}

func completenessPresent(value codegraph.Completeness) bool {
	return value.InputFiles != 0 || value.ParsedFiles != 0 || value.CheckedPackages != 0 || value.PartialPackages != 0 || value.ExactReferences != 0 || value.UnresolvedReferences != 0 || value.ExactCalls != 0 || value.UnresolvedCalls != 0 || value.UnresolvedImports != 0 || value.ExactReads != 0 || value.UnresolvedReads != 0 || value.ExactWrites != 0 || value.CandidateWrites != 0 || value.UnresolvedWrites != 0 || value.Truncated
}

func subsetOfKnown(values []string, known map[string]bool) bool {
	for _, value := range values {
		if !known[value] {
			return false
		}
	}
	return true
}

func subset(values, allowed []string) bool {
	set := map[string]bool{}
	for _, value := range allowed {
		set[value] = true
	}
	for _, value := range values {
		if !set[value] {
			return false
		}
	}
	return true
}

func diagnosticKey(diagnostic codegraph.Diagnostic) string {
	if diagnostic.Range == nil {
		return diagnostic.Code + "\x00" + diagnostic.Message
	}
	return fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%d", diagnostic.Code, diagnostic.Message, diagnostic.Range.FileID, diagnostic.Range.Start, diagnostic.Range.End)
}

func mapNodes(values map[string]codegraph.Node) []codegraph.Node {
	result := make([]codegraph.Node, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mapEdges(values map[string]codegraph.Edge) []codegraph.Edge {
	result := make([]codegraph.Edge, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func mapDiagnostics(values map[string]codegraph.Diagnostic) []codegraph.Diagnostic {
	result := make([]codegraph.Diagnostic, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return diagnosticKey(result[i]) < diagnosticKey(result[j]) })
	return result
}

func graphIsPartial(value codegraph.Completeness) bool {
	return value.Truncated || value.PartialPackages > 0
}
