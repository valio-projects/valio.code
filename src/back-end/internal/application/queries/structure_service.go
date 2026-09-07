package queries

import (
	"context"
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/structuregraph"
	"slices"
	"strings"
)

// Structure returns persisted, source-verified syntax graph neighborhoods.
func (s Service) Structure(ctx context.Context, q StructureQuery) (StructureResult, error) {
	result := StructureResult{Projection: structuregraph.SchemaVersion, Status: "unsupported", Files: []StructureFileResult{}, StopReasons: []string{}}
	if q.Scope.WorkspaceID != s.WorkspaceID {
		return result, fault.ErrForbidden
	}
	if q.Name == "" && q.FileID == "" && q.NodeID == "" || q.Name != "" && q.NodeID != "" || len(q.Name) > 4096 || q.Name != strings.TrimSpace(q.Name) || q.Depth < 0 || q.Depth > 12 || q.Limit < 0 || q.Limit > 2000 || len(q.Scope.ProjectIDs) > 128 {
		return result, fault.ErrInvalid
	}
	if q.Depth == 0 {
		q.Depth = 2
	}
	if q.Limit == 0 {
		q.Limit = 200
	}
	view, err := s.View(ctx, q.Scope.ViewID)
	if err != nil {
		return result, err
	}
	result.ViewID = view.ID
	if !viewHasProjects(view, q.Scope.ProjectIDs) {
		return result, fault.ErrNotFound
	}
	files := map[string]snapshots.FileRef{}
	for _, file := range view.Files {
		if len(q.Scope.ProjectIDs) == 0 || overlaps(file.ProjectIDs, q.Scope.ProjectIDs) {
			files[file.ID] = file
		}
	}
	if q.FileID != "" {
		if _, ok := files[q.FileID]; !ok {
			return result, fault.ErrNotFound
		}
	}
	if view.Projections["syntax_structure"] != "partial" && view.Projections["syntax_structure"] != "ready" {
		return result, nil
	}
	result.Status = "partial"
	artifacts, err := s.Store.Artifacts(ctx, view)
	if err != nil {
		return result, err
	}
	nodesLeft, edgesLeft := q.Limit, 5000
	for _, artifact := range artifacts {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		if artifact.ViewID != view.ID {
			return result, fault.ErrForbidden
		}
		file, ok := files[artifact.FileID]
		if !ok || q.FileID != "" && file.ID != q.FileID || len(artifact.StructureGraph) == 0 {
			continue
		}
		var graph structuregraph.Graph
		if json.Unmarshal(artifact.StructureGraph, &graph) != nil || graph.Validate() != nil || graph.Language != file.Language {
			return result, fault.ErrInvalid
		}
		for _, node := range graph.Nodes {
			if node.FileID != file.ID || node.RepositoryID != string(file.RepositoryID) || !slices.Equal(node.ProjectIDs, file.ProjectIDs) || node.Range.End > file.Size {
				return result, fault.ErrForbidden
			}
		}
		selected, reasons := selectStructure(ctx, graph, q, nodesLeft, edgesLeft)
		if len(selected.Nodes) == 0 {
			continue
		}
		selected.FileID = file.ID
		selected.Path = file.Path
		selected.Language = file.Language
		nodesLeft -= len(selected.Nodes)
		edgesLeft -= len(selected.Edges)
		result.Files = append(result.Files, selected)
		for _, reason := range reasons {
			if !slices.Contains(result.StopReasons, reason) {
				result.StopReasons = append(result.StopReasons, reason)
			}
		}
		if nodesLeft <= 0 || edgesLeft <= 0 {
			result.StopReasons = append(result.StopReasons, "OUTPUT_BUDGET")
			break
		}
	}
	result.Truncated = len(result.StopReasons) > 0
	return result, ctx.Err()
}
