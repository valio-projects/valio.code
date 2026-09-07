package retrieval

// Representation is a deterministic retrieval text view. Text is never an LLM
// summary; generated templates are derived only from facts present in the chunk.
type Representation struct {
	// Kind identifies the retrieval profile that produced Text.
	Kind RepresentationKind `json:"kind"`
	// Text is canonical source or a deterministic fact-derived template.
	Text string `json:"text"`
}
