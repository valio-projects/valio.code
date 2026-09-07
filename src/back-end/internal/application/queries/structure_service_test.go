package queries

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/valio-projects/valio.code/internal/application/fault"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
	"github.com/valio-projects/valio.code/internal/projects"
	"github.com/valio-projects/valio.code/internal/structuregraph"
)

func TestStructureReturnsNestedDeclarationMembership(t *testing.T) {
	view, artifacts, _ := structureFixture(t)
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	result, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, Name: "User", Depth: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "partial" || len(result.Files) != 1 || result.Truncated {
		t.Fatalf("structure result = %+v", result)
	}
	for _, name := range []string{"User", "Name", "Rename", "value"} {
		if !structureResultHasNode(result, name) {
			t.Fatalf("missing declaration membership node %q: %+v", name, result.Files[0].Nodes)
		}
	}
	if !structureResultHasEdge(result, structuregraph.RelationHasParameter) || !structureResultHasEdge(result, structuregraph.RelationDeclares) {
		t.Fatalf("missing nested declaration edges: %+v", result.Files[0].Edges)
	}
}

func TestStructureReportsDepthAndOutputLimits(t *testing.T) {
	view, artifacts, _ := structureFixture(t)
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	depth, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, Name: "User", Depth: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !depth.Truncated || !hasStopReason(depth, "DEPTH_BUDGET") || structureResultHasNode(depth, "value") {
		t.Fatalf("depth-one result = %+v", depth)
	}
	limited, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, Name: "User", Depth: 2, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !limited.Truncated || !hasStopReason(limited, "NODE_BUDGET") || len(limited.Files) != 1 || len(limited.Files[0].Nodes) != 1 {
		t.Fatalf("limited result = %+v", limited)
	}
}

func TestStructureDoesNotLeakAcrossProjectOrFileScope(t *testing.T) {
	view, artifacts, ids := structureFixture(t)
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	projectTwo, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p2"}}, Name: "User"})
	if err != nil {
		t.Fatal(err)
	}
	if len(projectTwo.Files) != 0 {
		t.Fatalf("project p2 received p1 declaration: %+v", projectTwo)
	}
	_, err = service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, FileID: ids.otherFile})
	if !errors.Is(err, fault.ErrNotFound) {
		t.Fatalf("out-of-scope file error = %v", err)
	}
}

func TestStructureRejectsForgedGraphMembership(t *testing.T) {
	for _, forge := range []struct {
		name  string
		apply func(*structuregraph.Graph)
	}{
		{name: "file", apply: func(graph *structuregraph.Graph) {
			for index := range graph.Nodes {
				graph.Nodes[index].FileID = "outside"
				graph.Nodes[index].Range.FileID = "outside"
				graph.Nodes[index].Evidence.Range.FileID = "outside"
			}
			for index := range graph.Edges {
				graph.Edges[index].Evidence.Range.FileID = "outside"
			}
		}},
		{name: "repository", apply: func(graph *structuregraph.Graph) {
			for index := range graph.Nodes {
				graph.Nodes[index].RepositoryID = "outside"
			}
		}},
		{name: "range", apply: func(graph *structuregraph.Graph) {
			graph.Nodes[0].Range.End += 1 << 20
			graph.Nodes[0].Evidence.Range.End += 1 << 20
		}},
	} {
		t.Run(forge.name, func(t *testing.T) {
			view, artifacts, _ := structureFixture(t)
			var graph structuregraph.Graph
			if err := json.Unmarshal(artifacts[0].StructureGraph, &graph); err != nil {
				t.Fatal(err)
			}
			forge.apply(&graph)
			artifacts[0].StructureGraph, _ = json.Marshal(graph)
			service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
			_, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, Name: "User"})
			if !errors.Is(err, fault.ErrForbidden) {
				t.Fatalf("forged %s graph error = %v", forge.name, err)
			}
		})
	}
}

func TestStructureReportsUnknownFileAndUnsupportedOldView(t *testing.T) {
	view, artifacts, _ := structureFixture(t)
	service := Service{Store: graphFixtureStore{view: view, artifacts: artifacts}, WorkspaceID: "workspace"}
	_, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, FileID: "missing"})
	if !errors.Is(err, fault.ErrNotFound) {
		t.Fatalf("unknown file error = %v", err)
	}
	delete(view.Projections, "syntax_structure")
	service.Store = graphFixtureStore{view: view, artifacts: artifacts}
	result, err := service.Structure(context.Background(), StructureQuery{Scope: SearchScope{WorkspaceID: "workspace", ViewID: view.ID, ProjectIDs: []string{"p1"}}, Name: "User"})
	if err != nil || result.Status != "unsupported" || len(result.Files) != 0 {
		t.Fatalf("old view result=%+v err=%v", result, err)
	}
}

type structureIDs struct {
	userFile  string
	otherFile string
}

func structureFixture(t *testing.T) (snapshots.View, []snapshots.Artifact, structureIDs) {
	t.Helper()
	content := "class User { string Name; void Rename(string value) {} }"
	userReport := syntaxfacts.Report{Schema: 1, Language: "csharp", Path: "User.cs", ValidSyntax: true, Capabilities: syntaxfacts.Capabilities{Syntax: true}}
	fieldStart := strings.Index(content, "string Name")
	methodStart := strings.Index(content, "void Rename")
	parameterStart := strings.Index(content, "string value")
	userReport.Symbols = []syntaxfacts.Declaration{
		{ID: "class", Name: "User", Kind: "class", Start: 0, End: len(content)},
		{ID: "field", ParentID: "class", Name: "Name", Kind: "field", Start: fieldStart, End: fieldStart + len("string Name;")},
		{ID: "method", ParentID: "class", Name: "Rename", Kind: "method", Start: methodStart, End: strings.Index(content[methodStart:], "}") + methodStart + 1, Parameters: []syntaxfacts.Parameter{{Name: "value", Start: parameterStart, End: parameterStart + len("string value")}}},
	}
	userGraph, err := structuregraph.Build(context.Background(), structuregraph.Source{ID: "user-file", RepositoryID: "repo", Path: "User.cs", Content: content, ProjectIDs: []string{"p1"}}, userReport)
	if err != nil {
		t.Fatal(err)
	}
	otherContent := "class Other {}"
	otherReport := syntaxfacts.Report{Schema: 1, Language: "csharp", Path: "Other.cs", ValidSyntax: true, Symbols: []syntaxfacts.Declaration{{ID: "class", Name: "Other", Kind: "class", Start: 0, End: len(otherContent)}}, Capabilities: syntaxfacts.Capabilities{Syntax: true}}
	otherGraph, err := structuregraph.Build(context.Background(), structuregraph.Source{ID: "other-file", RepositoryID: "repo", Path: "Other.cs", Content: otherContent, ProjectIDs: []string{"p2"}}, otherReport)
	if err != nil {
		t.Fatal(err)
	}
	userJSON, _ := json.Marshal(userGraph)
	otherJSON, _ := json.Marshal(otherGraph)
	view := snapshots.View{ID: "structure-view", WorkspaceID: domain.WorkspaceID("workspace"), Projects: []projects.Definition{{Project: domain.Project{ID: "p1"}}, {Project: domain.Project{ID: "p2"}}}, Files: []snapshots.FileRef{{ID: "user-file", RepositoryID: "repo", Path: "User.cs", Language: "csharp", Size: len(content), ProjectIDs: []string{"p1"}}, {ID: "other-file", RepositoryID: "repo", Path: "Other.cs", Language: "csharp", Size: len(otherContent), ProjectIDs: []string{"p2"}}}, Projections: map[string]string{"syntax_structure": "partial"}}
	artifacts := []snapshots.Artifact{{ID: "user-artifact", FileID: "user-file", ViewID: view.ID, StructureGraph: userJSON}, {ID: "other-artifact", FileID: "other-file", ViewID: view.ID, StructureGraph: otherJSON}}
	return view, artifacts, structureIDs{userFile: "user-file", otherFile: "other-file"}
}

func structureResultHasNode(result StructureResult, name string) bool {
	for _, file := range result.Files {
		for _, node := range file.Nodes {
			if node.Name == name {
				return true
			}
		}
	}
	return false
}

func structureResultHasEdge(result StructureResult, kind structuregraph.RelationKind) bool {
	for _, file := range result.Files {
		for _, edge := range file.Edges {
			if edge.Kind == kind {
				return true
			}
		}
	}
	return false
}

func hasStopReason(result StructureResult, expected string) bool {
	for _, reason := range result.StopReasons {
		if reason == expected {
			return true
		}
	}
	return false
}
