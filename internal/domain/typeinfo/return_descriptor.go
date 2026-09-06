package typeinfo

type ReturnDescriptor struct {
	ID         string         `json:"id"`
	Name       Fact[string]   `json:"name"`
	Position   Fact[int]      `json:"position"`
	Type       TypeReference  `json:"type"`
	Attributes []AttributeUse `json:"attributes"`
}
