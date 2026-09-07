package typeinfo

// TypeReference identifies a type use and its resolution evidence.
type TypeReference struct {
	// Name is the source or rendered type name when known.
	Name Fact[string] `json:"name"`
	// Symbol resolves the referenced type without promoting candidates to exact.
	Symbol SymbolReference `json:"symbol"`
	// Attributes lists annotations on this type use, including a return type.
	Attributes []AttributeUse `json:"attributes"`
	// Array supplies rank and bounds for array-shaped uses.
	Array *ArrayShape `json:"array,omitempty"`
}
