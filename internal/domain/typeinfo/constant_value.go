package typeinfo

type ConstantValue struct {
	Type  TypeReference `json:"type"`
	Value Fact[string]  `json:"value"` // exact compiler-computed text, not a float
}
