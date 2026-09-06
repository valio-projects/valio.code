package typeinfo

type NamedAttributeArgument struct {
	Name  Fact[string]   `json:"name"`
	Value AttributeValue `json:"value"`
}
