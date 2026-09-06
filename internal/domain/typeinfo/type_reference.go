package typeinfo

type TypeReference struct {
	Name       Fact[string]    `json:"name"`
	Symbol     SymbolReference `json:"symbol"`
	Attributes []AttributeUse  `json:"attributes"` // annotations on the type use, including a return type
	Array      *ArrayShape     `json:"array,omitempty"`
}
