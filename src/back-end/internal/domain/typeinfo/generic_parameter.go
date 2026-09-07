package typeinfo

// GenericParameter describes a declared type or method parameter.
type GenericParameter struct {
	// ID is stable within the containing declaration.
	ID string `json:"id"`
	// Name is the source parameter name when known.
	Name Fact[string] `json:"name"`
	// Constraints lists declared bounds or interfaces.
	Constraints []TypeReference `json:"constraints"`
	// Attributes lists annotations on the parameter.
	Attributes []AttributeUse `json:"attributes"`
}
