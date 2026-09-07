package queries

// RetrievalQuery selects a ranking operation over one pinned source view.
type RetrievalQuery struct {
	// Scope fixes workspace, project membership and source version.
	Scope SearchScope `json:"scope"`
	// Query is plain text, not the boolean source-search DSL.
	Query string `json:"query"`
	// Mode is lexical, symbol, structural, semantic or hybrid.
	Mode string `json:"mode"`
	// TargetChunkID selects the source example for structural search.
	TargetChunkID string `json:"targetChunkId,omitempty"`
	// ModelProfile selects an operator-configured AI profile for semantic search.
	ModelProfile string `json:"modelProfile,omitempty"`
	// Representation selects a vector representation kind, default code.
	Representation string `json:"representation,omitempty"`
	// Limit bounds returned hits, default 20, maximum 100.
	Limit int `json:"limit,omitempty"`
	// Rerank requests the configured cross encoder after candidate fusion.
	Rerank bool `json:"rerank,omitempty"`
}
