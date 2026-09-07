package retrieval

import (
	"context"
	"reflect"
	"sort"
	"strings"
	"testing"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/syntaxfacts"
)

func stringPointer(value string) *string { return &value }

func syntaxReport(language, path string, declarations ...syntaxfacts.Declaration) syntaxfacts.Report {
	return syntaxfacts.Report{Schema: 1, Language: language, Path: path, ValidSyntax: true, Symbols: declarations, Capabilities: syntaxfacts.Capabilities{Syntax: true}}
}

func span(content, fragment string) (int, int) {
	start := strings.Index(content, fragment)
	if start < 0 {
		panic("fixture fragment not found")
	}
	return start, start + len(fragment)
}

func TestBuildSyntaxPreservesUTF8NestedDeclarationsAndMetadata(t *testing.T) {
	content := "// Привет 😀\n[Attr] public class User {\n  public string Name;\n  public string Label(string prefix) { return prefix + Name; }\n}\n"
	classStart, classEnd := span(content, "[Attr] public class User {\n  public string Name;\n  public string Label(string prefix) { return prefix + Name; }\n}")
	fieldStart, fieldEnd := span(content, "public string Name;")
	methodStart, methodEnd := span(content, "public string Label(string prefix) { return prefix + Name; }")
	parameterStart, parameterEnd := span(content, "string prefix")
	report := syntaxReport("csharp", "src/user.cs",
		syntaxfacts.Declaration{ID: "user", Name: "User", Kind: "class", Start: classStart, End: classEnd, Visibility: "public", Attributes: []string{"[Attr]"}},
		syntaxfacts.Declaration{ID: "name", Name: "Name", Kind: "field", ParentID: "user", Start: fieldStart, End: fieldEnd, Type: stringPointer("string"), Visibility: "public"},
		syntaxfacts.Declaration{ID: "label", Name: "Label", Kind: "method", ParentID: "user", Start: methodStart, End: methodEnd, Type: stringPointer("string"), Visibility: "public", Parameters: []syntaxfacts.Parameter{{Name: "prefix", Type: stringPointer("string"), Start: parameterStart, End: parameterEnd}}},
	)
	source := Source{ID: "file", RepositoryID: "repo", Path: "src/user.cs", Language: "csharp", Content: content, ProjectIDs: []string{"project"}}
	chunks, err := (Builder{MaxBytes: 4096}).BuildSyntax(context.Background(), source, report)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) != 5 {
		t.Fatalf("chunks = %d, want source coverage plus direct field and method chunks", len(chunks))
	}
	classChunk := domainretrieval.Chunk{}
	labelChunk := domainretrieval.Chunk{}
	for _, chunk := range chunks {
		if chunk.ParentID == "" && chunk.Kind == domainretrieval.ChunkType {
			classChunk = chunk
		}
		if chunk.ParentID != "" && chunk.Kind == domainretrieval.ChunkMethod {
			labelChunk = chunk
		}
	}
	if classChunk.ID == "" || labelChunk.ID == "" {
		t.Fatalf("missing class or direct method chunk: %#v", chunks)
	}
	if classChunk.Kind != domainretrieval.ChunkType || classChunk.Text != content[classStart:classEnd] || classChunk.Start != classStart || classChunk.End != classEnd {
		t.Fatalf("class lost canonical UTF-8 range: %#v", classChunk)
	}
	if strings.Contains(classChunk.Header, classChunk.Text) || !strings.Contains(classChunk.Header, "syntax_kind=class") || !strings.Contains(classChunk.Header, "attributes=[Attr]") {
		t.Fatalf("syntax metadata header = %q", classChunk.Header)
	}
	foundSymbol := false
	for _, representation := range classChunk.Representations {
		if representation.Kind == domainretrieval.RepresentationSymbol && strings.Contains(representation.Text, "symbol: User") {
			foundSymbol = true
		}
	}
	if !foundSymbol {
		t.Fatal("syntax declaration did not produce a symbol representation")
	}
	if labelChunk.ParentID != classChunk.ID || labelChunk.Text != content[methodStart:methodEnd] {
		t.Fatalf("direct method lost its lexical parent or source range: %#v", labelChunk)
	}
	labelSymbol := false
	for _, representation := range labelChunk.Representations {
		if representation.Kind == domainretrieval.RepresentationSymbol && strings.Contains(representation.Text, "symbol: Label(prefix: string) -> string") {
			labelSymbol = true
		}
	}
	if !labelSymbol {
		t.Fatal("direct method did not produce its own symbol representation")
	}
	coverage := []domainretrieval.Chunk{}
	for _, chunk := range chunks {
		if chunk.ParentID == "" && chunk.Text != "" {
			coverage = append(coverage, chunk)
		}
	}
	sort.Slice(coverage, func(i, j int) bool { return coverage[i].Start < coverage[j].Start })
	var rebuilt strings.Builder
	for _, chunk := range coverage {
		rebuilt.WriteString(chunk.Text)
	}
	if rebuilt.String() != content {
		t.Fatal("syntax chunks did not preserve source coverage")
	}
}

func TestBuildSyntaxLargeDeclarationGroupsDirectChildrenAndCoversGaps(t *testing.T) {
	body := strings.Repeat("    work();\n", 20)
	content := "import { dep } from './dep';\n// keep this comment\nclass User {\n  field: string;\n  method(prefix: string): string {\n" + body + "    return prefix;\n  }\n}\n// trailing configuration\n"
	classStart, classEnd := span(content, "class User {\n  field: string;\n  method(prefix: string): string {\n"+body+"    return prefix;\n  }\n}")
	fieldStart, fieldEnd := span(content, "field: string;")
	methodStart, methodEnd := span(content, "method(prefix: string): string {\n"+body+"    return prefix;\n  }")
	parameterStart, parameterEnd := span(content, "prefix: string")
	report := syntaxReport("typescript", "src/user.ts",
		syntaxfacts.Declaration{ID: "user", Name: "User", Kind: "class", Start: classStart, End: classEnd},
		syntaxfacts.Declaration{ID: "field", Name: "field", Kind: "field", ParentID: "user", Start: fieldStart, End: fieldEnd, Type: stringPointer("string")},
		syntaxfacts.Declaration{ID: "method", Name: "method", Kind: "method", ParentID: "user", Start: methodStart, End: methodEnd, Type: stringPointer("string"), Parameters: []syntaxfacts.Parameter{{Name: "prefix", Type: stringPointer("string"), Start: parameterStart, End: parameterEnd}}},
	)
	source := Source{ID: "file", RepositoryID: "repo", Path: "src/user.ts", Language: "typescript", Content: content, ProjectIDs: []string{"project"}}
	builder := Builder{MaxBytes: 64}
	first, err := builder.BuildSyntax(context.Background(), source, report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := builder.BuildSyntax(context.Background(), source, report)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("syntax build is not deterministic: %v", err)
	}
	classParent := domainretrieval.Chunk{}
	for _, chunk := range first {
		if chunk.Kind == domainretrieval.ChunkType && chunk.ParentID == "" && chunk.Text == "" {
			classParent = chunk
		}
	}
	if classParent.ID == "" {
		t.Fatal("large class grouping parent missing")
	}
	if classParent.Kind != domainretrieval.ChunkType || !classParent.Split || classParent.Text != "" {
		t.Fatalf("large class did not become grouping parent: %#v", classParent)
	}
	methodParent := domainretrieval.Chunk{}
	for _, chunk := range first {
		if chunk.ParentID == classParent.ID && chunk.Kind == domainretrieval.ChunkMethod && chunk.Text == "" {
			methodParent = chunk
		}
	}
	if methodParent.ID == "" {
		t.Fatal("large direct method did not retain its class grouping parent")
	}
	for _, chunk := range first {
		if chunk.Text != "" && (len(chunk.Text) > 64 || chunk.Text != content[chunk.Start:chunk.End]) {
			t.Fatalf("invalid bounded canonical syntax chunk: %#v", chunk)
		}
	}
	textChunks := append([]domainretrieval.Chunk(nil), first...)
	sort.Slice(textChunks, func(i, j int) bool { return textChunks[i].Start < textChunks[j].Start })
	position := 0
	for _, chunk := range textChunks {
		if chunk.Text == "" {
			continue
		}
		if chunk.Start != position {
			t.Fatalf("source gap at %d before %#v", position, chunk)
		}
		position = chunk.End
	}
	if position != len(content) {
		t.Fatalf("source coverage ended at %d, want %d", position, len(content))
	}
}

func TestBuildSyntaxRejectsMalformedDeclarationRelationships(t *testing.T) {
	source := Source{ID: "file", RepositoryID: "repo", Path: "src/user.ts", Language: "typescript", Content: "class User {}"}
	validStart, validEnd := span(source.Content, "class User {}")
	for _, declarations := range [][]syntaxfacts.Declaration{
		{{ID: "one", Name: "one", Kind: "class", Start: validStart, End: validEnd}, {ID: "two", Name: "two", Kind: "method", Start: validStart + 1, End: validEnd}},
		{{ID: "one", Name: "one", Kind: "class", ParentID: "two", Start: validStart, End: validEnd}, {ID: "two", Name: "two", Kind: "method", ParentID: "one", Start: validStart, End: validEnd}},
		{{ID: "one", Name: "one", Kind: "class", Start: validStart, End: validEnd + 1}},
	} {
		if _, err := (Builder{}).BuildSyntax(context.Background(), source, syntaxReport("typescript", source.Path, declarations...)); err == nil {
			t.Fatal("malformed syntax report built chunks")
		}
	}
}

func TestBuildSyntaxOrdersEqualSpansByParentDepth(t *testing.T) {
	source := Source{ID: "file", RepositoryID: "repo", Path: "src/user.ts", Language: "typescript", Content: "class User {}"}
	start, end := span(source.Content, "class User {}")
	chunks, err := (Builder{}).BuildSyntax(context.Background(), source, syntaxReport("typescript", source.Path,
		syntaxfacts.Declaration{ID: "s10", Name: "User", Kind: "class", Start: start, End: end},
		syntaxfacts.Declaration{ID: "s9", Name: "wrapped", Kind: "method", ParentID: "s10", Start: start, End: end},
	))
	if err != nil || len(chunks) != 2 || chunks[1].ParentID != chunks[0].ID {
		t.Fatalf("equal-span syntax wrapper did not retain hierarchy: %#v %v", chunks, err)
	}
}

// TestBuildSyntaxHelperFixturesPreserveCanonicalRanges keeps the builder aligned
// with the byte-based report contract emitted by the Tree-sitter helper. In
// particular, TypeScript type annotations retain their written whitespace.
func TestBuildSyntaxHelperFixturesPreserveCanonicalRanges(t *testing.T) {
	fixtures := []struct {
		name, language, path, content string
		declarations                  []syntaxfacts.Declaration
	}{
		{
			name:     "C# property",
			language: "csharp",
			path:     "user.cs",
			content:  "public class User { public int ID { get; set; } }",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "class", Start: 0, End: 49},
				{ID: "s2", Name: "ID", Kind: "property", ParentID: "s1", Start: 20, End: 47},
			},
		},
		{
			name:     "C struct field",
			language: "c",
			path:     "user.c",
			content:  "struct User { int id; };",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "struct", Start: 0, End: 23},
				{ID: "s2", Name: "id", Kind: "field", ParentID: "s1", Start: 14, End: 21},
			},
		},
		{
			name:     "C plus plus field",
			language: "cpp",
			path:     "user.cpp",
			content:  "class User { public: int id; };",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "class", Start: 0, End: 30},
				{ID: "s2", Name: "id", Kind: "field", ParentID: "s1", Start: 21, End: 28},
			},
		},
		{
			name:     "JavaScript method",
			language: "javascript",
			path:     "user.js",
			content:  "export class User { get() { return 1; } }",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "class", Start: 7, End: 41},
				{ID: "s2", Name: "get", Kind: "method", ParentID: "s1", Start: 20, End: 39},
			},
		},
		{
			name:     "TypeScript field spacing",
			language: "typescript",
			path:     "user.ts",
			content:  "export class User { id: number; }",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "class", Start: 7, End: 33},
				{ID: "s2", Name: "id", Kind: "field", ParentID: "s1", Start: 20, End: 30},
			},
		},
		{
			name:     "Java field",
			language: "java",
			path:     "User.java",
			content:  "public class User { public int id; }",
			declarations: []syntaxfacts.Declaration{
				{ID: "s1", Name: "User", Kind: "class", Start: 0, End: 36},
				{ID: "s2", Name: "id", Kind: "field", ParentID: "s1", Start: 20, End: 34},
			},
		},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			source := Source{ID: "file", RepositoryID: "repo", Path: fixture.path, Language: fixture.language, Content: fixture.content, ProjectIDs: []string{"project"}}
			chunks, err := (Builder{}).BuildSyntax(context.Background(), source, syntaxReport(fixture.language, fixture.path, fixture.declarations...))
			if err != nil {
				t.Fatal(err)
			}
			for _, chunk := range chunks {
				if chunk.Text != "" && (len(chunk.Text) != chunk.End-chunk.Start || chunk.Text != fixture.content[chunk.Start:chunk.End]) {
					t.Fatalf("chunk lost its canonical source range: %#v", chunk)
				}
			}
		})
	}
}

func TestBuildSyntaxHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	source := Source{ID: "file", RepositoryID: "repo", Path: "src/user.ts", Language: "typescript", Content: "class User {}"}
	start, end := span(source.Content, "class User {}")
	if _, err := (Builder{}).BuildSyntax(ctx, source, syntaxReport("typescript", source.Path, syntaxfacts.Declaration{ID: "user", Name: "User", Kind: "class", Start: start, End: end})); err == nil {
		t.Fatal("cancelled syntax build succeeded")
	}
}
