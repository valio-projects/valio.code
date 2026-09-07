package typeinfo

// ReturnDescriptor describes one callable result.
type ReturnDescriptor struct {
	// ID is stable within the containing callable.
	ID string `json:"id"`
	// Name is the result name when the language permits one.
	Name Fact[string] `json:"name"`
	// Position is the zero-based result position.
	Position Fact[int] `json:"position"`
	// Type identifies the returned value type.
	Type TypeReference `json:"type"`
	// Attributes lists annotations on this result.
	Attributes []AttributeUse `json:"attributes"`
}
