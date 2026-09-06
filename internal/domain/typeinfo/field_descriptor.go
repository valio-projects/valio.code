package typeinfo

// FieldDescriptor describes a declared field without asserting physical layout.
type FieldDescriptor struct {
	// ID is stable within the containing type descriptor.
	ID string `json:"id"`
	// Name is the source field name when known.
	Name Fact[string] `json:"name"`
	// Type identifies the declared field type.
	Type TypeReference `json:"type"`
	// Visibility records access scope when the language exposes it.
	Visibility Fact[Visibility] `json:"visibility"`
	// Modifiers preserves language-level field flags.
	Modifiers Fact[[]Modifier] `json:"modifiers"`
	// Attributes lists declaration annotations.
	Attributes []AttributeUse `json:"attributes"`
	// Tag is a Go struct tag, not a CLR attribute.
	Tag Fact[string] `json:"tag"`
	// Declaration identifies the defining source occurrence when known.
	Declaration *OccurrenceLink `json:"declaration,omitempty"`
}
