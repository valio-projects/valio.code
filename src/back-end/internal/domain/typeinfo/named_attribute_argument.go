package typeinfo

// NamedAttributeArgument binds an explicit argument name to its retained value.
type NamedAttributeArgument struct {
	// Name is the source argument name when known.
	Name Fact[string] `json:"name"`
	// Value contains the literal, expression, or redaction record.
	Value AttributeValue `json:"value"`
}
