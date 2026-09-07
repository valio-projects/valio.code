package retrieval

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

func build(t *testing.T, max int, sources ...Source) []domainretrieval.Chunk {
	t.Helper()
	chunks, err := (Builder{MaxBytes: max}).Build(context.Background(), sources)
	if err != nil {
		t.Fatal(err)
	}
	return chunks
}

func TestBuildGoChunksKeepCanonicalRangesAndRepresentations(t *testing.T) {
	content := "package sample\n\n// PaymentService sends invoices.\ntype PaymentService struct{}\n\n// Pay returns an error when payment fails.\nfunc Pay() error {\n\treturn nil\n}\n"
	source := Source{ID: "file", RepositoryID: "repo", Path: "payment.go", Language: "go", Content: content, ProjectIDs: []string{"payments", "billing", "payments"}}
	chunks := build(t, 4096, source)
	if len(chunks) != 2 {
		t.Fatalf("chunks = %d, want type and function", len(chunks))
	}
	for _, chunk := range chunks {
		if chunk.Text != content[chunk.Start:chunk.End] {
			t.Fatalf("chunk %s lost canonical source range", chunk.ID)
		}
		if strings.Contains(chunk.Header, chunk.Text) || !strings.Contains(chunk.Header, "language=go") {
			t.Fatalf("header is not separate metadata: %q", chunk.Header)
		}
		if !reflect.DeepEqual(chunk.ProjectIDs, []string{"billing", "payments"}) {
			t.Fatalf("project scope = %#v", chunk.ProjectIDs)
		}
	}
	pay := chunks[1]
	if pay.Kind != domainretrieval.ChunkFunction {
		t.Fatalf("kind = %q", pay.Kind)
	}
	kinds := map[domainretrieval.RepresentationKind]bool{}
	for _, representation := range pay.Representations {
		kinds[representation.Kind] = true
	}
	for _, kind := range []domainretrieval.RepresentationKind{domainretrieval.RepresentationCode, domainretrieval.RepresentationSymbol, domainretrieval.RepresentationDocumentation, domainretrieval.RepresentationError, domainretrieval.RepresentationAPI} {
		if !kinds[kind] {
			t.Errorf("missing representation %q", kind)
		}
	}
}

func TestOversizedFunctionUsesBoundedChildren(t *testing.T) {
	content := "package sample\nfunc Huge() {\n" + strings.Repeat("work()\n", 50) + "}\n"
	chunks := build(t, 64, Source{ID: "file", RepositoryID: "repo", Path: "huge.go", Language: "go", Content: content})
	if len(chunks) < 2 {
		t.Fatal("oversized function was not split")
	}
	parent := chunks[0]
	if parent.Kind != domainretrieval.ChunkFunction || !parent.Split || parent.Text != "" {
		t.Fatalf("unexpected grouping parent: %#v", parent)
	}
	for _, chunk := range chunks[1:] {
		if chunk.ParentID != parent.ID || !chunk.Split {
			t.Fatalf("child does not retain parent relation: %#v", chunk)
		}
		if len(chunk.Text) > 64 || chunk.Text != content[chunk.Start:chunk.End] {
			t.Fatalf("child exceeds budget or lost source range: %#v", chunk)
		}
	}
}

func TestFallbackChunksRespectUTF8AndAreDeterministic(t *testing.T) {
	content := strings.Repeat("αβγ", 40)
	first := Source{ID: "b", RepositoryID: "repo", Path: "b.txt", Language: "text", Content: content, ProjectIDs: []string{"p"}}
	second := Source{ID: "a", RepositoryID: "repo", Path: "a.txt", Language: "text", Content: content, ProjectIDs: []string{"p"}}
	chunks := build(t, 64, first)
	var restored strings.Builder
	for _, chunk := range chunks {
		if chunk.Kind != domainretrieval.ChunkFallback || !chunk.Fallback || len(chunk.Text) > 64 || !utf8.ValidString(chunk.Text) {
			t.Fatalf("invalid UTF-8 fallback: %#v", chunk)
		}
		restored.WriteString(chunk.Text)
	}
	if restored.String() != content {
		t.Fatal("fallback chunks do not reconstruct source")
	}
	ordered := build(t, 64, first, second)
	reversed := build(t, 64, second, first)
	if !reflect.DeepEqual(ordered, reversed) {
		t.Fatal("chunk build depends on input order")
	}
}

func TestAbsentFactsDoNotCreateSyntheticViews(t *testing.T) {
	chunks := build(t, 4096, Source{ID: "file", RepositoryID: "repo", Path: "plain.go", Language: "go", Content: "package sample\nfunc private() {}\n"})
	if len(chunks) != 1 {
		t.Fatal("missing function chunk")
	}
	for _, representation := range chunks[0].Representations {
		if representation.Kind == domainretrieval.RepresentationDocumentation || representation.Kind == domainretrieval.RepresentationError || representation.Kind == domainretrieval.RepresentationAPI || representation.Kind == domainretrieval.RepresentationArchitecture || representation.Kind == domainretrieval.RepresentationChange {
			t.Fatalf("created unsupported representation: %#v", representation)
		}
	}
}

func TestBuilderValidatesBudgetIdentityPathAndUTF8(t *testing.T) {
	valid := Source{ID: "file", RepositoryID: "one", Path: "valid.txt", Language: "text", Content: "valid"}
	if _, err := (Builder{MaxBytes: 63}).Build(context.Background(), []Source{valid}); err == nil {
		t.Fatal("accepted undersized chunk budget")
	}
	duplicate := valid
	duplicate.RepositoryID = "two"
	if _, err := (Builder{}).Build(context.Background(), []Source{valid, duplicate}); err == nil {
		t.Fatal("accepted duplicate file ID across repositories")
	}
	invalidPath := valid
	invalidPath.ID, invalidPath.Path = "path", "../outside.txt"
	if _, err := (Builder{}).Build(context.Background(), []Source{invalidPath}); err == nil {
		t.Fatal("accepted unsafe path")
	}
	invalidUTF8 := valid
	invalidUTF8.ID, invalidUTF8.Content = "utf8", string([]byte{0xff})
	if _, err := (Builder{}).Build(context.Background(), []Source{invalidUTF8}); err == nil {
		t.Fatal("accepted invalid UTF-8")
	}
}

func TestBuilderHonorsCancelledContextBeforeOversizedFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := (Builder{MaxBytes: 64}).Build(ctx, []Source{{ID: "file", RepositoryID: "repo", Path: "large.txt", Language: "text", Content: strings.Repeat("x", 10000)}})
	if err == nil {
		t.Fatal("cancelled build succeeded")
	}
}
