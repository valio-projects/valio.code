package structuregraph

// Evidence identifies the helper observation that supports a graph fact.
type Evidence struct {
	// Producer identifies the helper and its parser-level observation.
	Producer string `json:"producer"`
	// Range identifies the observed source span when it exists.
	Range *ByteRange `json:"range,omitempty"`
}
