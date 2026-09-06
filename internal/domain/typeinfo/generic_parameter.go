package typeinfo

type GenericParameter struct {
	ID          string          `json:"id"`
	Name        Fact[string]    `json:"name"`
	Constraints []TypeReference `json:"constraints"`
	Attributes  []AttributeUse  `json:"attributes"`
}
