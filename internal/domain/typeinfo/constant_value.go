package typeinfo

// ConstantValue preserves a constant's type and exact textual value.
type ConstantValue struct {
	// Type identifies the constant's declared or resolved type.
	Type TypeReference `json:"type"`
	// Value is exact compiler-computed text, never a lossy floating-point value.
	Value Fact[string] `json:"value"`
}
