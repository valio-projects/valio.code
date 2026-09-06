package typeinfo

// TypeLayout describes a particular target/build profile. Source declaration
// order, CLR attributes or host pointer width alone cannot establish offsets.
type TypeLayout struct {
	// TargetArchitecture is the target CPU or platform identity for physical facts.
	TargetArchitecture Fact[string] `json:"targetArchitecture"`
	// ABI is the applicable application binary interface.
	ABI Fact[string] `json:"abi"`
	// Compiler identifies the producer used to establish layout facts.
	Compiler Fact[string] `json:"compiler"`
	// CompilerVersion pins the compiler version used for layout facts.
	CompilerVersion Fact[string] `json:"compilerVersion"`
	// Kind identifies the target's placement policy.
	Kind Fact[LayoutKind] `json:"kind"`
	// SizeBytes is the complete type size in bytes when established.
	SizeBytes Fact[uint64] `json:"sizeBytes"`
	// AlignmentBytes is the target alignment requirement in bytes.
	AlignmentBytes Fact[uint64] `json:"alignmentBytes"`
	// PackingBytes is the configured packing boundary in bytes.
	PackingBytes Fact[uint64] `json:"packingBytes"`
	// Fields contains physical information for declared fields.
	Fields []FieldLayout `json:"fields"`
	// UnknownReasons records why no reliable physical value is available.
	UnknownReasons []string `json:"unknownReasons"`
}

// UnresolvedLayout constructs an empty layout whose facts consistently explain why.
func UnresolvedLayout(scope TypeScope, reason string) TypeLayout {
	return TypeLayout{TargetArchitecture: UnresolvedFact[string](scope, reason), ABI: UnresolvedFact[string](scope, reason), Compiler: UnresolvedFact[string](scope, reason), CompilerVersion: UnresolvedFact[string](scope, reason), Kind: UnresolvedFact[LayoutKind](scope, reason), SizeBytes: UnresolvedFact[uint64](scope, reason), AlignmentBytes: UnresolvedFact[uint64](scope, reason), PackingBytes: UnresolvedFact[uint64](scope, reason), Fields: []FieldLayout{}, UnknownReasons: []string{reason}}
}
