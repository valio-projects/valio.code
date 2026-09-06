package typeinfo

// PropertyDescriptor describes a property, including optional accessors.
type PropertyDescriptor struct {
	// ID is stable within the containing type descriptor.
	ID string `json:"id"`
	// Name is the declared property name.
	Name Fact[string] `json:"name"`
	// Type identifies the property value type.
	Type TypeReference `json:"type"`
	// Visibility records access scope when known.
	Visibility Fact[Visibility] `json:"visibility"`
	// Modifiers preserves language-specific property flags.
	Modifiers Fact[[]Modifier] `json:"modifiers"`
	// Parameters contains indexer parameters where supported.
	Parameters []ParameterDescriptor `json:"parameters"`
	// Getter describes the read accessor when present.
	Getter *MethodDescriptor `json:"getter,omitempty"`
	// Setter describes the write accessor when present.
	Setter *MethodDescriptor `json:"setter,omitempty"`
	// Attributes lists declaration annotations.
	Attributes []AttributeUse `json:"attributes"`
	// Declaration identifies the defining occurrence when known.
	Declaration *OccurrenceLink `json:"declaration,omitempty"`
}
