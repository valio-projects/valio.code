package retrieval

import (
	"strings"
	"testing"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

func contextChunk(id, parent, project, text string, start, end int) domainretrieval.Chunk {
	return domainretrieval.Chunk{ID: id, ParentID: parent, FileID: "file", RepositoryID: "repo", ProjectIDs: []string{project}, Start: start, End: end, Text: text, Header: "kind=test"}
}

func TestContextBuilderExpandsOnlyCallerScopedRelations(t *testing.T) {
	root := contextChunk("parent", "", "one", "", 0, 20)
	child := contextChunk("child", "parent", "one", "child", 0, 5)
	otherProject := contextChunk("other", "parent", "two", "other", 6, 11)
	context, err := (ContextBuilder{}).Build(root.ID, []domainretrieval.Chunk{root, child, otherProject}, ContextOptions{ViewID: "view", MaxBytes: 8000, Profile: domainretrieval.RepresentationCode})
	if err != nil {
		t.Fatal(err)
	}
	if len(context.Items) != 2 || context.Items[0].ChunkID != "parent" || context.Items[1].ChunkID != "child" {
		t.Fatalf("unexpected context items: %#v", context.Items)
	}
	if len(context.Omitted) != 1 || context.Omitted[0].ChunkID != "other" || context.Omitted[0].Reason != OmissionProjectScope {
		t.Fatalf("project crossing was not omitted: %#v", context.Omitted)
	}
}

func TestContextBuilderReportsBudgetOmissionAndKeepsChild(t *testing.T) {
	parent := contextChunk("parent", "", "one", strings.Repeat("x", 8000), 0, 8000)
	child := contextChunk("child", "parent", "one", "child", 0, 5)
	context, err := (ContextBuilder{}).Build(parent.ID, []domainretrieval.Chunk{parent, child}, ContextOptions{MaxBytes: 8000})
	if err != nil {
		t.Fatal(err)
	}
	if len(context.Items) != 1 || context.Items[0].ChunkID != child.ID {
		t.Fatalf("small child was not retained: %#v", context.Items)
	}
	if len(context.Omitted) != 1 || context.Omitted[0].ChunkID != parent.ID || context.Omitted[0].Reason != OmissionBudget {
		t.Fatalf("missing budget reason: %#v", context.Omitted)
	}
}
