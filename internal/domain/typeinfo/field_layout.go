package typeinfo

// FieldLayout records physical placement of one declared field for a build target.
type FieldLayout struct {
	// FieldID identifies a field in the owning descriptor.
	FieldID string `json:"fieldId"`
	// OffsetBytes is the byte offset from the enclosing object or record start.
	OffsetBytes Fact[uint64] `json:"offsetBytes"`
	// BitOffset is the bit offset within a packed storage unit when applicable.
	BitOffset Fact[uint64] `json:"bitOffset"`
	// BitWidth is the field width in bits when bit packing is used.
	BitWidth Fact[uint64] `json:"bitWidth"`
}
