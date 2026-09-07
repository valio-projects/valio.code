package queries

// EmbeddingCommand indexes a bounded page of persisted, policy-approved chunks.
// No caller-provided source text or provider URL is accepted.
type EmbeddingCommand struct {
	Scope          SearchScope `json:"scope"`
	ModelProfile   string      `json:"modelProfile"`
	Representation string      `json:"representation,omitempty"`
	Offset         int         `json:"offset,omitempty"`
	Limit          int         `json:"limit,omitempty"`
}
