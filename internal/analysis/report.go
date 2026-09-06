package analysis

type Report struct {
	Path            string          `json:"path"`
	Package         string          `json:"package"`
	ValidSyntax     bool            `json:"validSyntax"`
	Symbols         []Symbol        `json:"symbols"`
	Occurrences     []Occurrence    `json:"occurrences"`
	Imports         []Import        `json:"imports"`
	Diagnostics     []Diagnostic    `json:"diagnostics"`
	Capabilities    Capabilities    `json:"capabilities"`
	Types           []TypeInfo      `json:"types"`
	Functions       []FunctionInfo  `json:"functions"`
	ConstantGroups  []ConstantGroup `json:"constantGroups"`
	TypeCheckStatus string          `json:"typeCheckStatus"`
	TypeDiagnostics []Diagnostic    `json:"typeDiagnostics"`
}
