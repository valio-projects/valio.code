package typeinfo

// MethodDescriptor describes a callable declaration and its source evidence.
type MethodDescriptor struct {
	// ID is stable within the containing type descriptor.
	ID string `json:"id"`
	// Name is the declared callable name.
	Name Fact[string] `json:"name"`
	// Signature preserves a producer-rendered signature; it is not normalized ABI data.
	Signature Fact[string] `json:"signature"`
	// Visibility records access scope when known.
	Visibility Fact[Visibility] `json:"visibility"`
	// Modifiers preserves language-specific callable flags.
	Modifiers Fact[[]Modifier] `json:"modifiers"`
	// Receiver is the receiver parameter for languages that expose one.
	Receiver *ParameterDescriptor `json:"receiver,omitempty"`
	// Parameters lists inputs in declaration order.
	Parameters []ParameterDescriptor `json:"parameters"`
	// Returns lists outputs in declaration order.
	Returns []ReturnDescriptor `json:"returns"`
	// GenericParameters lists callable type parameters.
	GenericParameters []GenericParameter `json:"genericParameters"`
	// Attributes lists declaration annotations.
	Attributes []AttributeUse `json:"attributes"`
	// Declaration identifies the defining occurrence when known.
	Declaration *OccurrenceLink `json:"declaration,omitempty"`
}
