package retrieval

import domainretrieval "github.com/valio-projects/valio.code/internal/domain/retrieval"

// ContextItem is a cited canonical source record selected for bounded context.
type ContextItem struct {
	// ChunkID identifies the selected retrieval chunk.
	ChunkID string `json:"chunkId"`
	// FileID identifies the immutable source file.
	FileID string `json:"fileId"`
	// Start is the included original UTF-8 byte offset.
	Start int `json:"start"`
	// End is the excluded original UTF-8 byte offset.
	End int `json:"end"`
	// Text is the canonical source text, never a generated summary.
	Text string `json:"text"`
	// Header is generated metadata supplied separately from Text.
	Header string `json:"header"`
	// ParentID identifies the source grouping parent when present.
	ParentID string `json:"parentId,omitempty"`
	// Profile identifies the representation requested for this context assembly.
	Profile domainretrieval.RepresentationKind `json:"profile"`
}
