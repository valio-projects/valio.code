package structuregraph

import (
	"context"
	"strings"
	"testing"

	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

func TestBuildPreservesNestedDeclarationsAcrossSupportedSyntaxReports(t *testing.T) {
	for _, fixture := range []struct {
		language  string
		path      string
		content   string
		parent    string
		child     string
		method    string
		parameter string
	}{
		{language: "cpp", path: "box.cpp", content: "class Box { int value; void Set(int next); };", parent: "Box", child: "value", method: "Set", parameter: "next"},
		{language: "csharp", path: "Box.cs", content: "class Box { int Value; void Set(int next) {} }", parent: "Box", child: "Value", method: "Set", parameter: "next"},
		{language: "typescript", path: "box.ts", content: "class Box { value: number; set(next: number): void {} }", parent: "Box", child: "value", method: "set", parameter: "next"},
	} {
		t.Run(fixture.language, func(t *testing.T) {
			report := baseReport(fixture.language, fixture.path)
			report.Symbols = []syntaxfacts.Declaration{
				{ID: "s1", Name: fixture.parent, Kind: "class", Start: 0, End: len(fixture.content)},
				declaration("s2", "s1", fixture.child, "field", fixture.content, fixture.child),
				{ID: "s3", ParentID: "s1", Name: fixture.method, Kind: "method", Start: 0, End: len(fixture.content)},
			}
			method := &report.Symbols[2]
			method.Parameters = []syntaxfacts.Parameter{parameter(fixture.parameter, fixture.content)}
			result, err := Build(context.Background(), source("file-"+fixture.language, "repo", fixture.path, fixture.content, "project"), report)
			if err != nil {
				t.Fatal(err)
			}
			if err := result.Validate(); err != nil {
				t.Fatal(err)
			}
			parentID := nodeID(result, NodeDeclaration, fixture.parent)
			childID := nodeID(result, NodeDeclaration, fixture.child)
			methodID := nodeID(result, NodeDeclaration, fixture.method)
			parameterID := nodeID(result, NodeParameter, fixture.parameter)
			if parentID == "" || childID == "" || methodID == "" || parameterID == "" {
				t.Fatalf("missing nested graph nodes: %+v", result.Nodes)
			}
			if !hasEdge(result, RelationDeclares, parentID, childID, ResolutionExact) || !hasEdge(result, RelationDeclares, parentID, methodID, ResolutionExact) || !hasEdge(result, RelationHasParameter, methodID, parameterID, ResolutionExact) {
				t.Fatalf("missing exact lexical membership edges: %+v", result.Edges)
			}
		})
	}
}

func TestBuildKeepsImportsAndCallsUnresolved(t *testing.T) {
	content := "import { save } from \"./save\"; save();"
	report := baseReport("typescript", "use.ts")
	report.Imports = []syntaxfacts.Import{{Path: "./save", Kind: "import", Resolution: "unresolved", Start: 0, End: strings.Index(content, ";") + 1}}
	callStart := strings.LastIndex(content, "save")
	report.References = []syntaxfacts.Reference{{Name: "save", Kind: "call", Resolution: "unresolved", Start: callStart, End: callStart + len("save")}}
	result, err := Build(context.Background(), source("use", "repo", "use.ts", content, "project"), report)
	if err != nil {
		t.Fatal(err)
	}
	importID := nodeID(result, NodeImport, "./save")
	if importID == "" || nodeID(result, NodeCall, "save") == "" {
		t.Fatalf("missing unresolved observations: %+v", result.Nodes)
	}
	if !hasEdge(result, RelationImports, importID, "", ResolutionUnresolved) {
		t.Fatalf("missing unresolved import edge: %+v", result.Edges)
	}
	for _, edge := range result.Edges {
		if edge.Kind != RelationContains && edge.Kind != RelationDeclares && edge.Kind != RelationHasParameter && edge.Kind != RelationImports {
			t.Fatalf("unexpected semantic relation: %+v", edge)
		}
	}
}

func TestBuildDoesNotMergeSameNameAcrossProjectContexts(t *testing.T) {
	content := "class Same {}"
	report := baseReport("typescript", "same.ts")
	report.Symbols = []syntaxfacts.Declaration{declaration("s1", "", "Same", "class", content, "Same")}
	first, err := Build(context.Background(), source("file-one", "repo", "same.ts", content, "project-one"), report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Build(context.Background(), source("file-two", "repo", "same.ts", content, "project-two"), report)
	if err != nil {
		t.Fatal(err)
	}
	firstID, secondID := nodeID(first, NodeDeclaration, "Same"), nodeID(second, NodeDeclaration, "Same")
	if firstID == "" || secondID == "" || firstID == secondID {
		t.Fatalf("same source spelling merged across projects: %q %q", firstID, secondID)
	}
	if nodeByID(first, firstID).ProjectIDs[0] != "project-one" || nodeByID(second, secondID).ProjectIDs[0] != "project-two" {
		t.Fatal("project context was not retained on graph nodes")
	}
}

func TestBuildRejectsOutOfSourceRange(t *testing.T) {
	content := "class C {}"
	report := baseReport("typescript", "bad.ts")
	report.Symbols = []syntaxfacts.Declaration{{ID: "s1", Name: "C", Kind: "class", Start: 0, End: len(content) + 1}}
	if _, err := Build(context.Background(), source("bad", "repo", "bad.ts", content, "project"), report); err == nil {
		t.Fatal("accepted report range beyond supplied UTF-8 content")
	}
}

func baseReport(language, path string) syntaxfacts.Report {
	return syntaxfacts.Report{Schema: 1, Language: language, Path: path, ValidSyntax: true, Capabilities: syntaxfacts.Capabilities{Syntax: true, SemanticResolution: false}}
}

func source(id, repositoryID, path, content, projectID string) Source {
	return Source{ID: id, RepositoryID: repositoryID, Path: path, Content: content, ProjectIDs: []string{projectID}}
}

func declaration(id, parentID, name, kind, content, needle string) syntaxfacts.Declaration {
	start := strings.Index(content, needle)
	return syntaxfacts.Declaration{ID: id, ParentID: parentID, Name: name, Kind: kind, Start: start, End: start + len(needle)}
}

func parameter(name, content string) syntaxfacts.Parameter {
	start := strings.LastIndex(content, name)
	return syntaxfacts.Parameter{Name: name, Start: start, End: start + len(name)}
}

func nodeID(graph Graph, kind NodeKind, name string) string {
	for _, node := range graph.Nodes {
		if node.Kind == kind && node.Name == name {
			return node.ID
		}
	}
	return ""
}

func nodeByID(graph Graph, id string) Node {
	for _, node := range graph.Nodes {
		if node.ID == id {
			return node
		}
	}
	return Node{}
}

func hasEdge(graph Graph, kind RelationKind, sourceID, targetID string, resolution Resolution) bool {
	for _, edge := range graph.Edges {
		if edge.Kind == kind && edge.SourceID == sourceID && edge.TargetID == targetID && edge.Resolution == resolution {
			return true
		}
	}
	return false
}
