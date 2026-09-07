package retrieval

import (
	"strings"

	domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/structure"
)

// StructuralSimilarity compares two Go chunks using the existing normalized Go
// AST fingerprint. A match is a structural candidate, never proof that behavior
// or meaning is equivalent. Non-Go or invalid syntax returns zero similarity.
func StructuralSimilarity(left, right domainretrieval.Chunk) float64 {
	leftFingerprint, ok := structuralFingerprint(left)
	if !ok {
		return 0
	}
	rightFingerprint, ok := structuralFingerprint(right)
	if !ok || leftFingerprint.Hash != rightFingerprint.Hash {
		return 0
	}
	return 1
}

func structuralFingerprint(chunk domainretrieval.Chunk) (structure.Result, bool) {
	if chunk.Text == "" || !strings.Contains(chunk.Header, "language=go") {
		return structure.Result{}, false
	}
	source := chunk.Text
	if chunk.Kind == domainretrieval.ChunkStatement || chunk.Kind == domainretrieval.ChunkFallback {
		source = "package retrieval\nfunc candidate() {\n" + source + "\n}"
	} else if !strings.HasPrefix(strings.TrimSpace(source), "package ") {
		source = "package retrieval\n" + source
	}
	result, err := structure.Fingerprint(source)
	return result, err == nil
}
