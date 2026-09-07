package graph

import (
	"context"
	"strings"
	"testing"

	"github.com/valio-projects/valio.code/internal/domain/codegraph"
)

func TestBuildResolvesCrossFileCallsMembersTypesAndConstants(t *testing.T) {
	first := `package sample
type Record struct { Value int }
const Ready = 1
func (Record) Method() int { return Ready }
func First(r Record) int { return r.Value }
`
	second := `package sample
func Second() int { return First(Record{Value: Ready}) }
`
	result, err := Build(context.Background(), []Source{
		{ID: "a", RepositoryID: "repo", Path: "pkg/a.go", Content: first, ProjectIDs: []string{"project"}},
		{ID: "b", RepositoryID: "repo", Path: "pkg/b.go", Content: second, ProjectIDs: []string{"project"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
	if result.Completeness.ExactCalls != 1 {
		t.Fatalf("exact calls = %d, want 1", result.Completeness.ExactCalls)
	}
	if result.Completeness.ExactReferences < 4 {
		t.Fatalf("exact references = %d, want local cross-file references", result.Completeness.ExactReferences)
	}
	if !hasNode(result, codegraph.NodeSymbol, "Ready", "a") || !hasNode(result, codegraph.NodeSymbol, "Value", "a") {
		t.Fatal("constant and field symbols were not recorded")
	}
	if !hasRelation(result, codegraph.RelationCalls, codegraph.ResolutionExact) {
		t.Fatal("missing exact call relation")
	}
	if !hasCallableContainment(result, "Second", "b") {
		t.Fatal("missing parser-evidenced callable-to-call-site containment")
	}
	if !hasRelation(result, codegraph.RelationMemberOf, codegraph.ResolutionExact) {
		t.Fatal("missing exact method or field member relation")
	}
	if !hasRelation(result, codegraph.RelationHasType, codegraph.ResolutionExact) {
		t.Fatal("missing exact local type relation")
	}
	for _, node := range result.Nodes {
		if strings.Contains(string(node.Kind), "enum") {
			t.Fatal("Go constants must not be represented as enums")
		}
	}
}

func TestBuildResolvesShadowedUseToLocalDefinition(t *testing.T) {
	content := `package sample
var Value = 1
func Use() int {
	Value := 2
	return Value
}
`
	result, err := Build(context.Background(), []Source{{ID: "shadow", RepositoryID: "repo", Path: "shadow.go", Content: content, ProjectIDs: []string{"project"}}})
	if err != nil {
		t.Fatal(err)
	}
	localOffset := strings.Index(content, "Value :=")
	returnOffset := strings.LastIndex(content, "Value")
	localSymbol := ""
	returnReference := ""
	for _, node := range result.Nodes {
		if node.FileID != "shadow" || node.Range == nil || node.Name != "Value" {
			continue
		}
		if node.Kind == codegraph.NodeSymbol && node.Range.Start == localOffset {
			localSymbol = node.ID
		}
		if node.Kind == codegraph.NodeReference && node.Range.Start == returnOffset {
			returnReference = node.ID
		}
	}
	if localSymbol == "" || returnReference == "" {
		t.Fatalf("shadow symbols missing: local=%q reference=%q", localSymbol, returnReference)
	}
	for _, edge := range result.Edges {
		if edge.Kind == codegraph.RelationRefersTo && edge.SourceID == returnReference && edge.Resolution == codegraph.ResolutionExact {
			if edge.TargetID != localSymbol {
				t.Fatalf("shadowed use resolved to %q, want %q", edge.TargetID, localSymbol)
			}
			return
		}
	}
	t.Fatal("missing exact reference for shadowed local")
}

func TestBuildDoesNotCombineProjectContextsOrDirectories(t *testing.T) {
	projectOne := `package same
func Only() {}
`
	projectTwo := `package same
func Use() { Only() }
`
	differentDirectory := `package same
func Other() { Only() }
`
	result, err := Build(context.Background(), []Source{
		{ID: "one", RepositoryID: "repo", Path: "same/one.go", Content: projectOne, ProjectIDs: []string{"one"}},
		{ID: "two", RepositoryID: "repo", Path: "same/two.go", Content: projectTwo, ProjectIDs: []string{"two"}},
		{ID: "other", RepositoryID: "repo", Path: "other/other.go", Content: differentDirectory, ProjectIDs: []string{"two"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, edge := range result.Edges {
		if edge.Resolution != codegraph.ResolutionExact || len(edge.ProjectIDs) != 1 || edge.ProjectIDs[0] != "two" {
			continue
		}
		target := nodeByID(result, edge.TargetID)
		if target != nil && target.FileID == "one" {
			t.Fatalf("exact relation crossed incompatible project or directory: %+v", edge)
		}
	}
	if result.Completeness.CheckedPackages != 3 {
		t.Fatalf("checked package variants = %d, want 3", result.Completeness.CheckedPackages)
	}
}

func TestBuildReportsUnresolvedImportsAndSeparatesCallsFromReferences(t *testing.T) {
	content := `package sample
import "fmt"
func Use() { fmt.Println("x") }
`
	result, err := Build(context.Background(), []Source{{ID: "import", RepositoryID: "repo", Path: "use.go", Content: content, ProjectIDs: []string{"project"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Completeness.UnresolvedImports != 1 {
		t.Fatalf("unresolved imports = %d, want 1", result.Completeness.UnresolvedImports)
	}
	if result.Completeness.UnresolvedCalls != 1 {
		t.Fatalf("unresolved calls = %d, want 1", result.Completeness.UnresolvedCalls)
	}
	if !hasRelation(result, codegraph.RelationImports, codegraph.ResolutionUnresolved) || !hasRelation(result, codegraph.RelationCalls, codegraph.ResolutionUnresolved) {
		t.Fatal("unresolved import or call relation missing")
	}
	if !hasNode(result, codegraph.NodeReference, "Println", "import") || !hasNode(result, codegraph.NodeCallSite, "call", "import") {
		t.Fatal("call site and identifier reference must be distinct nodes")
	}
	if !hasDiagnostic(result, "unresolved_import") {
		t.Fatal("missing unresolved import diagnostic")
	}
}

func TestBuildDoesNotTreatConversionsOrBuiltinsAsExactCalls(t *testing.T) {
	content := `package sample
func Use() int { return len([]int{int('x')}) }
`
	result, err := Build(context.Background(), []Source{{ID: "builtins", RepositoryID: "repo", Path: "builtins.go", Content: content, ProjectIDs: []string{"project"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Completeness.ExactCalls != 0 || result.Completeness.UnresolvedCalls != 2 {
		t.Fatalf("exact=%d unresolved=%d; conversions and builtins have no local function target", result.Completeness.ExactCalls, result.Completeness.UnresolvedCalls)
	}
}

func TestBuildRecordsExactFieldReadsAndWritesSeparatelyFromCalls(t *testing.T) {
	content := `package sample
type User struct { ID int }
func Getter(user User) int { return user.ID }
func Update(user *User) { user.ID += 1; user.ID++ }
func Call(user User) int { return Getter(user) }
`
	result, err := Build(context.Background(), []Source{{ID: "access", RepositoryID: "repo", Path: "access.go", Content: content, ProjectIDs: []string{"project"}}})
	if err != nil {
		t.Fatal(err)
	}
	fieldSymbol := symbolAt(result, "access", strings.Index(content, "ID int"))
	if fieldSymbol == "" {
		t.Fatal("field symbol missing")
	}
	if got := relationCountTo(result, codegraph.RelationReads, codegraph.ResolutionExact, fieldSymbol); got != 3 {
		t.Fatalf("exact field reads = %d, want getter plus compound and increment reads", got)
	}
	if got := relationCountTo(result, codegraph.RelationWrites, codegraph.ResolutionExact, fieldSymbol); got != 2 {
		t.Fatalf("exact field writes = %d, want compound and increment writes", got)
	}
	if result.Completeness.ExactCalls != 1 {
		t.Fatalf("calls were conflated with field access: exact calls = %d", result.Completeness.ExactCalls)
	}
}

func TestBuildRecordsInitializationAndConservativeArrayPointerWrites(t *testing.T) {
	content := `package sample
const Ready = 1
func Use(items []int, pointer *int) int {
	value := Ready
	value += 1
	items[0] = value
	*pointer = value
	return Ready
}
`
	result, err := Build(context.Background(), []Source{{ID: "write-kinds", RepositoryID: "repo", Path: "write_kinds.go", Content: content, ProjectIDs: []string{"project"}}})
	if err != nil {
		t.Fatal(err)
	}
	valueSymbol := symbolAt(result, "write-kinds", strings.Index(content, "value :="))
	readySymbol := symbolAt(result, "write-kinds", strings.Index(content, "Ready ="))
	if valueSymbol == "" || readySymbol == "" {
		t.Fatal("variable or constant symbol missing")
	}
	if got := relationCountTo(result, codegraph.RelationWrites, codegraph.ResolutionExact, valueSymbol); got != 2 {
		t.Fatalf("exact writes = %d, want initialization and compound assignment", got)
	}
	if got := relationCountTo(result, codegraph.RelationWrites, codegraph.ResolutionExact, readySymbol); got != 0 {
		t.Fatalf("constant declaration must not be represented as a write: %d", got)
	}
	if got := relationCountTo(result, codegraph.RelationReads, codegraph.ResolutionExact, readySymbol); got != 2 {
		t.Fatalf("constant reads = %d, want initializer and return", got)
	}
	if result.Completeness.CandidateWrites != 1 || result.Completeness.UnresolvedWrites != 1 {
		t.Fatalf("array/pointer writes candidate=%d unresolved=%d", result.Completeness.CandidateWrites, result.Completeness.UnresolvedWrites)
	}
}

func TestBuildRejectsDuplicateIDsAndHonorsCancellation(t *testing.T) {
	_, err := Build(context.Background(), []Source{{ID: "same", RepositoryID: "repo", Path: "a.go", Content: "package p"}, {ID: "same", RepositoryID: "repo", Path: "b.go", Content: "package p"}})
	if err == nil {
		t.Fatal("duplicate source IDs accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Build(ctx, []Source{{ID: "a", RepositoryID: "repo", Path: "a.go", Content: "package p"}})
	if err == nil {
		t.Fatal("cancelled context accepted")
	}
}

func hasNode(graph codegraph.Graph, kind codegraph.NodeKind, name, fileID string) bool {
	for _, node := range graph.Nodes {
		if node.Kind == kind && node.Name == name && node.FileID == fileID {
			return true
		}
	}
	return false
}

func hasRelation(graph codegraph.Graph, kind codegraph.RelationKind, resolution codegraph.Resolution) bool {
	for _, edge := range graph.Edges {
		if edge.Kind == kind && edge.Resolution == resolution {
			return true
		}
	}
	return false
}

func hasDiagnostic(graph codegraph.Graph, code string) bool {
	for _, diagnostic := range graph.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func nodeByID(graph codegraph.Graph, id string) *codegraph.Node {
	for index := range graph.Nodes {
		if graph.Nodes[index].ID == id {
			return &graph.Nodes[index]
		}
	}
	return nil
}

func hasCallableContainment(graph codegraph.Graph, name, fileID string) bool {
	for _, edge := range graph.Edges {
		if edge.Kind != codegraph.RelationContains || edge.Resolution != codegraph.ResolutionExact {
			continue
		}
		source, target := nodeByID(graph, edge.SourceID), nodeByID(graph, edge.TargetID)
		if source != nil && target != nil && source.Kind == codegraph.NodeSymbol && source.Name == name && source.FileID == fileID && target.Kind == codegraph.NodeCallSite {
			return true
		}
	}
	return false
}

func symbolAt(graph codegraph.Graph, fileID string, offset int) string {
	for _, node := range graph.Nodes {
		if node.Kind == codegraph.NodeSymbol && node.FileID == fileID && node.Range != nil && node.Range.Start == offset {
			return node.ID
		}
	}
	return ""
}

func relationCountTo(graph codegraph.Graph, kind codegraph.RelationKind, resolution codegraph.Resolution, targetID string) int {
	count := 0
	for _, edge := range graph.Edges {
		if edge.Kind == kind && edge.Resolution == resolution && edge.TargetID == targetID {
			count++
		}
	}
	return count
}
