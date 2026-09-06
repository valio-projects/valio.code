package typeinfo

// ParameterDescriptor describes one callable input or receiver.
type ParameterDescriptor struct {
	// ID is stable within the containing callable.
	ID string `json:"id"`
	// Name is the declared name when present.
	Name Fact[string] `json:"name"`
	// Position is the zero-based declaration position.
	Position Fact[int] `json:"position"`
	// Type identifies the parameter type.
	Type TypeReference `json:"type"`
	// Modifiers preserves flags such as variadic or readonly.
	Modifiers Fact[[]Modifier] `json:"modifiers"`
	// DefaultValue is source text when a default exists and was collected.
	DefaultValue Fact[string] `json:"defaultValue"`
	// Attributes lists annotations on this parameter.
	Attributes []AttributeUse `json:"attributes"`
}
