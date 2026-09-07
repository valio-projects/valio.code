package retrieval

// Chunk is a bounded retrieval record anchored to immutable canonical source.
// Header is generated metadata and Text remains the exact UTF-8 source slice
// described by Start and End. A grouping parent for split source has empty Text
// and is never a text-search candidate; its children carry the source text.
type Chunk struct {
	// ID deterministically identifies this chunk's file, range, kind, and parent.
	ID string `json:"id"`
	// FileID identifies the immutable source file supplied to the builder.
	FileID string `json:"fileId"`
	// RepositoryID scopes the source repository.
	RepositoryID string `json:"repositoryId"`
	// ProjectIDs contains sorted, unique project memberships for this source file.
	ProjectIDs []string `json:"projectIds"`
	// Start is the included byte offset in the original UTF-8 source.
	Start int `json:"start"`
	// End is the excluded byte offset in the original UTF-8 source.
	End int `json:"end"`
	// Text is the canonical source slice, empty only for a split grouping parent.
	Text string `json:"text"`
	// Header is deterministic generated metadata separate from Text.
	Header string `json:"header"`
	// ParentID identifies the enclosing grouping chunk when this chunk is a child.
	ParentID string `json:"parentId,omitempty"`
	// Kind identifies the syntax or fallback boundary.
	Kind ChunkKind `json:"kind"`
	// Split reports that this record participates in an oversized-source split.
	Split bool `json:"split"`
	// Fallback reports that raw text splitting was required after syntax partitioning.
	Fallback bool `json:"fallback"`
	// Representations contains deterministic views supported by present source facts.
	Representations []Representation `json:"representations"`
}
