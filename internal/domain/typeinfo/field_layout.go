package typeinfo

type FieldLayout struct {
	FieldID     string       `json:"fieldId"`
	OffsetBytes Fact[uint64] `json:"offsetBytes"`
	BitOffset   Fact[uint64] `json:"bitOffset"`
	BitWidth    Fact[uint64] `json:"bitWidth"`
}
