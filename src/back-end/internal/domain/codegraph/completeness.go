package codegraph

// Completeness reports what the bounded local builder actually observed.
type Completeness struct {
	// InputFiles is the number of accepted Source inputs.
	InputFiles int `json:"inputFiles"`
	// ParsedFiles is the number of files for which go/parser returned an AST.
	ParsedFiles int `json:"parsedFiles"`
	// CheckedPackages is the number of repository/directory/package/project variants checked by go/types.
	CheckedPackages int `json:"checkedPackages"`
	// PartialPackages is the number of groups with parser or type-check diagnostics.
	PartialPackages int `json:"partialPackages"`
	// ExactReferences counts local reference relations established by go/types.
	ExactReferences int `json:"exactReferences"`
	// UnresolvedReferences counts identifier use nodes without a local exact target.
	UnresolvedReferences int `json:"unresolvedReferences"`
	// ExactCalls counts call relations established by go/types.
	ExactCalls int `json:"exactCalls"`
	// UnresolvedCalls counts call sites without a local exact function/method target.
	UnresolvedCalls int `json:"unresolvedCalls"`
	// UnresolvedImports counts written imports not resolved from supplied local inputs.
	UnresolvedImports int `json:"unresolvedImports"`
	// ExactReads counts local readable targets established by go/types.
	ExactReads int `json:"exactReads"`
	// UnresolvedReads counts read occurrences without a supplied local target.
	UnresolvedReads int `json:"unresolvedReads"`
	// ExactWrites counts local write targets established by go/types.
	ExactWrites int `json:"exactWrites"`
	// CandidateWrites counts conservative array-base write candidates.
	CandidateWrites int `json:"candidateWrites"`
	// UnresolvedWrites counts writes that cannot claim a supplied local target.
	UnresolvedWrites int `json:"unresolvedWrites"`
	// Truncated reports that a builder limit stopped traversal before all facts were emitted.
	Truncated bool `json:"truncated"`
}
