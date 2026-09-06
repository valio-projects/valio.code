package typeinfo

// TypeLayout describes a particular target/build profile. Source declaration
// order, CLR attributes or host pointer width alone cannot establish offsets.
type TypeLayout struct {
	TargetArchitecture Fact[string]     `json:"targetArchitecture"`
	ABI                Fact[string]     `json:"abi"`
	Compiler           Fact[string]     `json:"compiler"`
	CompilerVersion    Fact[string]     `json:"compilerVersion"`
	Kind               Fact[LayoutKind] `json:"kind"`
	SizeBytes          Fact[uint64]     `json:"sizeBytes"`
	AlignmentBytes     Fact[uint64]     `json:"alignmentBytes"`
	PackingBytes       Fact[uint64]     `json:"packingBytes"`
	Fields             []FieldLayout    `json:"fields"`
	UnknownReasons     []string         `json:"unknownReasons"`
}

func UnresolvedLayout(scope TypeScope, reason string) TypeLayout {
	return TypeLayout{TargetArchitecture: UnresolvedFact[string](scope, reason), ABI: UnresolvedFact[string](scope, reason), Compiler: UnresolvedFact[string](scope, reason), CompilerVersion: UnresolvedFact[string](scope, reason), Kind: UnresolvedFact[LayoutKind](scope, reason), SizeBytes: UnresolvedFact[uint64](scope, reason), AlignmentBytes: UnresolvedFact[uint64](scope, reason), PackingBytes: UnresolvedFact[uint64](scope, reason), Fields: []FieldLayout{}, UnknownReasons: []string{reason}}
}
