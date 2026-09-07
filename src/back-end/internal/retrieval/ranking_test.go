package retrieval

import (
	"math"
	"testing"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
)

func rankedChunk(id, project, text string) domainretrieval.Chunk {
	return domainretrieval.Chunk{ID: id, FileID: id, RepositoryID: "repo", ProjectIDs: []string{project}, Text: text, Header: "language=go; path=" + id}
}

func TestLexicalRankSplitsIdentifiersAndHonorsScope(t *testing.T) {
	chunks := []domainretrieval.Chunk{rankedChunk("payment", "billing", "type PaymentService struct{}"), rankedChunk("other", "other", "type Service struct{}")}
	result := LexicalRank("payment service", chunks, RankOptions{ProjectIDs: []string{"billing"}})
	if len(result) != 1 || result[0].Chunk.ID != "payment" || result[0].Provenance[0].Channel != ChannelLexical || result[0].Provenance[0].Rank != 1 {
		t.Fatalf("unexpected lexical result: %#v", result)
	}
	if result := LexicalRank("missing", chunks, RankOptions{}); len(result) != 0 {
		t.Fatalf("negative query returned %#v", result)
	}
}

func TestRankBM25RejectsOversizeScopeAndNonFiniteConfiguration(t *testing.T) {
	chunks := []domainretrieval.Chunk{rankedChunk("payment", "billing", "type PaymentService struct{}")}
	if _, err := RankBM25("payment", chunks, BM25Options{MaxQueryBytes: 2}); err == nil {
		t.Fatal("accepted query beyond configured limit")
	}
	if _, err := RankBM25("payment", chunks, BM25Options{K1: math.Inf(1)}); err == nil {
		t.Fatal("accepted infinite BM25 configuration")
	}
	projects := make([]string, 129)
	for i := range projects {
		projects[i] = "p" + string(rune('a'+i%26)) + string(rune('A'+i/26))
	}
	if _, err := RankBM25("payment", chunks, BM25Options{Scope: RankOptions{ProjectIDs: projects}}); err == nil {
		t.Fatal("accepted oversized scope")
	}
}

func TestFuseRRFUsesDefaultKAndKeepsChannelProvenance(t *testing.T) {
	chunk := rankedChunk("one", "p", "payment")
	lexical := []ScoredChunk{{Chunk: chunk, Score: 2, Provenance: []ChannelScore{{Channel: ChannelLexical, Score: 2, Rank: 1}}}}
	structural := []ScoredChunk{{Chunk: chunk, Score: 1, Provenance: []ChannelScore{{Channel: ChannelStructural, Score: 1, Rank: 2}}}}
	result := FuseRRF(0, lexical, structural)
	want := 1.0/61 + 1.0/62
	if len(result) != 1 || math.Abs(result[0].Score-want) > 1e-12 || len(result[0].Provenance) != 2 {
		t.Fatalf("unexpected RRF result: %#v", result)
	}
}

func TestStructuralSimilarityUsesNormalizedGoSyntax(t *testing.T) {
	left := domainretrieval.Chunk{Kind: domainretrieval.ChunkFunction, Header: "language=go", Text: "func F() int { return 1 }"}
	right := domainretrieval.Chunk{Kind: domainretrieval.ChunkFunction, Header: "language=go", Text: "func F() int { return 2 }"}
	if got := StructuralSimilarity(left, right); got != 1 {
		t.Fatalf("structural similarity = %v", got)
	}
	if got := StructuralSimilarity(left, domainretrieval.Chunk{Header: "language=text", Text: "func F() int { return 2 }"}); got != 0 {
		t.Fatalf("non-Go similarity = %v", got)
	}
}
