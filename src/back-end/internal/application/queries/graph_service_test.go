package queries

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/codegraph"
	"github.com/valio-projects/valio.code/internal/domain/repositories"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/projects"
)

// graphFixtureStore embeds the unused publication port so individual graph
// tests only need to implement the three read methods that Graph invokes.
type graphFixtureStore struct {
	repositories.SnapshotRepository
	view      snapshots.View
	artifacts []snapshots.Artifact
}

func (s graphFixtureStore) Latest(context.Context) (snapshots.View, error) {
	return s.view, nil
}

func (s graphFixtureStore) View(_ context.Context, id string) (snapshots.View, error) {
	if id != "" && id != s.view.ID {
		return snapshots.View{}, fault.ErrNotFound
	}
	return s.view, nil
}

func (s graphFixtureStore) Artifacts(context.Context, snapshots.View) ([]snapshots.Artifact, error) {
	return s.artifacts, nil
}

func TestGraphMergesCrossFileShardsAndScopesReferences(t *testing.T) {
	view, artifacts, ids := graphFixture()
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	result, err := service.Graph(context.Background(), GraphQuery{Mode: GraphReferences, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, TargetNodeID: ids.symbol})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != GraphComplete || len(result.Edges) != 1 || result.Edges[0].Kind != codegraph.RelationRefersTo {
		t.Fatalf("unexpected scoped graph result: %+v", result)
	}
	if !resultHasNode(result, ids.symbol) || !resultHasNode(result, ids.reference) {
		t.Fatal("cross-file reference and target were not merged")
	}
	result, err = service.Graph(context.Background(), GraphQuery{Mode: GraphReferences, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p2"}}, TargetNodeID: ids.symbol})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Nodes) != 0 || len(result.Edges) != 0 {
		t.Fatal("project scope exposed p1 graph facts")
	}
	if result.Status != GraphPartial || len(result.Notices) != 1 || result.Notices[0].Code != GraphMissingShard {
		t.Fatalf("missing p2 shard was not scoped explicitly: %+v", result)
	}
}

func TestGraphFindsCallersAndBoundedPaths(t *testing.T) {
	view, artifacts, ids := graphFixture()
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	callers, err := service.Graph(context.Background(), GraphQuery{Mode: GraphCallers, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, TargetNodeID: ids.symbol})
	if err != nil {
		t.Fatal(err)
	}
	if len(callers.Edges) != 1 || callers.Edges[0].Kind != codegraph.RelationCalls {
		t.Fatalf("callers = %+v", callers.Edges)
	}
	callees, err := service.Graph(context.Background(), GraphQuery{Mode: GraphCallees, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, TargetNodeID: ids.owner})
	if err != nil {
		t.Fatal(err)
	}
	if !resultHasNode(callees, ids.call) || !resultHasNode(callees, ids.symbol) || len(callees.Edges) != 2 {
		t.Fatalf("callees = %+v", callees)
	}
	paths, err := service.Graph(context.Background(), GraphQuery{Mode: GraphPaths, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, SourceNodeID: ids.call, TargetNodeID: ids.symbol, Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths.Paths) != 1 || len(paths.Paths[0]) != 2 || paths.Paths[0][1] != ids.symbol {
		t.Fatalf("paths = %+v", paths.Paths)
	}
}

func TestGraphFindsReadsAndWritesByTarget(t *testing.T) {
	view, artifacts, ids := graphFixture()
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	for _, mode := range []GraphMode{GraphReads, GraphWrites} {
		result, err := service.Graph(context.Background(), GraphQuery{Mode: mode, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, TargetNodeID: ids.symbol})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Nodes) != 2 || len(result.Edges) != 1 || result.Edges[0].Kind != codegraph.RelationKind(mode) {
			t.Fatalf("%s result = %+v", mode, result)
		}
	}
}

func TestGraphReportsUnsupportedOldViewAndRejectsShardLeak(t *testing.T) {
	view, artifacts, ids := graphFixture()
	service := Service{Store: graphFixtureStore{view: view, artifacts: []snapshots.Artifact{{ID: "old", FileID: "file-a", ViewID: view.ID}}}, WorkspaceID: "workspace"}
	result, err := service.Graph(context.Background(), GraphQuery{Mode: GraphSymbols, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, Name: "Target"})
	if err != nil || result.Status != GraphUnsupported || len(result.Notices) != 1 || result.Notices[0].Code != GraphUnsupportedOldView {
		t.Fatalf("old view result=%+v err=%v", result, err)
	}
	leaking := artifacts[0]
	var shard codegraph.Graph
	if err := json.Unmarshal(leaking.Graph, &shard); err != nil {
		t.Fatal(err)
	}
	shard.Nodes[0].FileID = "outside-view"
	leaking.Graph, _ = json.Marshal(shard)
	service.Store = graphFixtureStore{view: view, artifacts: []snapshots.Artifact{leaking, artifacts[1]}}
	_, err = service.Graph(context.Background(), GraphQuery{Mode: GraphReferences, Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID}, TargetNodeID: ids.symbol})
	if !errors.Is(err, fault.ErrForbidden) {
		t.Fatalf("cross-view graph node error = %v", err)
	}
}

func TestGraphQueryBoundsAndWorkspaceBoundary(t *testing.T) {
	if DefaultGraphDepth != 4 || MaxGraphDepth != 12 || DefaultGraphTargets != 100 || MaxGraphTargets != 1000 {
		t.Fatalf("unexpected graph limits: depth=%d/%d targets=%d/%d", DefaultGraphDepth, MaxGraphDepth, DefaultGraphTargets, MaxGraphTargets)
	}
	service := Service{WorkspaceID: "workspace"}
	valid := GraphQuery{Mode: GraphNeighbors, Scope: SearchScope{WorkspaceID: "workspace"}, TargetNodeID: "node", Depth: MaxGraphDepth, Limit: MaxGraphTargets}
	if err := validateGraphQuery(service, valid); err != nil {
		t.Fatalf("maximum valid query: %v", err)
	}
	valid.Depth = MaxGraphDepth + 1
	if !errors.Is(validateGraphQuery(service, valid), fault.ErrInvalid) {
		t.Fatal("depth above maximum was accepted")
	}
	valid.Depth = 0
	valid.Limit = MaxGraphTargets + 1
	if !errors.Is(validateGraphQuery(service, valid), fault.ErrInvalid) {
		t.Fatal("target limit above maximum was accepted")
	}
	valid.Limit = 0
	valid.Scope.WorkspaceID = "other"
	if !errors.Is(validateGraphQuery(service, valid), fault.ErrForbidden) {
		t.Fatal("workspace crossing was not forbidden")
	}
}

type graphIDs struct {
	symbol    string
	reference string
	call      string
	owner     string
}

func graphFixture() (snapshots.View, []snapshots.Artifact, graphIDs) {
	view := snapshots.View{
		ID:          "view",
		WorkspaceID: "workspace",
		Projects:    []projects.Definition{{Project: domain.Project{ID: "p1"}}, {Project: domain.Project{ID: "p2"}}},
		Files: []snapshots.FileRef{
			{ID: "file-a", RepositoryID: "repository", Path: "a.go", Language: "go", ProjectIDs: []string{"p1"}},
			{ID: "file-b", RepositoryID: "repository", Path: "b.go", Language: "go", ProjectIDs: []string{"p1"}},
			{ID: "file-c", RepositoryID: "repository", Path: "c.go", Language: "go", ProjectIDs: []string{"p2"}},
		},
	}
	ids := graphIDs{symbol: "symbol", reference: "reference", call: "call", owner: "owner"}
	symbol := codegraph.Node{ID: ids.symbol, Kind: codegraph.NodeSymbol, Name: "Target", FileID: "file-a", RepositoryID: "repository", ProjectIDs: []string{"p1"}}
	reference := codegraph.Node{ID: ids.reference, Kind: codegraph.NodeReference, Name: "Target", FileID: "file-b", RepositoryID: "repository", ProjectIDs: []string{"p1"}}
	call := codegraph.Node{ID: ids.call, Kind: codegraph.NodeCallSite, Name: "call", FileID: "file-b", RepositoryID: "repository", ProjectIDs: []string{"p1"}}
	owner := codegraph.Node{ID: ids.owner, Kind: codegraph.NodeSymbol, Name: "Caller", FileID: "file-b", RepositoryID: "repository", ProjectIDs: []string{"p1"}}
	first := codegraph.Graph{Schema: codegraph.SchemaVersion, Nodes: []codegraph.Node{symbol}, Completeness: codegraph.Completeness{InputFiles: 2, ParsedFiles: 2, CheckedPackages: 1}}
	second := codegraph.Graph{Schema: codegraph.SchemaVersion, Nodes: []codegraph.Node{reference, call, owner}, Edges: []codegraph.Edge{
		{ID: "reference-edge", Kind: codegraph.RelationRefersTo, SourceID: ids.reference, TargetID: ids.symbol, Resolution: codegraph.ResolutionExact, ProjectIDs: []string{"p1"}},
		{ID: "call-edge", Kind: codegraph.RelationCalls, SourceID: ids.call, TargetID: ids.symbol, Resolution: codegraph.ResolutionExact, ProjectIDs: []string{"p1"}},
		{ID: "contains-call", Kind: codegraph.RelationContains, SourceID: ids.owner, TargetID: ids.call, Resolution: codegraph.ResolutionExact, ProjectIDs: []string{"p1"}},
		{ID: "read-edge", Kind: codegraph.RelationReads, SourceID: ids.reference, TargetID: ids.symbol, Resolution: codegraph.ResolutionExact, ProjectIDs: []string{"p1"}},
		{ID: "write-edge", Kind: codegraph.RelationWrites, SourceID: ids.reference, TargetID: ids.symbol, Resolution: codegraph.ResolutionExact, ProjectIDs: []string{"p1"}},
	}}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	return view, []snapshots.Artifact{{ID: "artifact-a", FileID: "file-a", ViewID: view.ID, Graph: firstJSON}, {ID: "artifact-b", FileID: "file-b", ViewID: view.ID, Graph: secondJSON}}, ids
}

func resultHasNode(result GraphResult, id string) bool {
	for _, node := range result.Nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}
