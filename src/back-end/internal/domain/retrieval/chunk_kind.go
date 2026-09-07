// Package retrieval defines immutable, source-grounded retrieval records.
package retrieval

// ChunkKind identifies the source boundary represented by a Chunk.
type ChunkKind string

const (
	// ChunkFileFallback identifies a deterministic text fragment for a non-Go file
	// or syntax that could not be partitioned safely.
	ChunkFileFallback ChunkKind = "file_fallback"
	// ChunkType identifies a Go type declaration grouping record.
	ChunkType ChunkKind = "type"
	// ChunkFunction identifies a Go function declaration grouping record.
	ChunkFunction ChunkKind = "function"
	// ChunkMethod identifies a Go method declaration grouping record.
	ChunkMethod ChunkKind = "method"
	// ChunkStatement identifies a Go statement selected from an oversized body.
	ChunkStatement ChunkKind = "statement"
	// ChunkFallback identifies a bounded raw-source fragment whose syntax boundary
	// exceeded the configured size budget.
	ChunkFallback ChunkKind = "fallback"
)
