package codegraph

// Graph is a bounded, in-memory result. Its IDs do not contain an analysis view
// version and are only stable while supplied Source IDs and source ranges remain stable.
type Graph struct {
	// Schema identifies the encoding and builder contract.
	Schema string `json:"schema"`
	// Nodes is deterministic by ID.
	Nodes []Node `json:"nodes"`
	// Edges is deterministic by ID.
	Edges []Edge `json:"edges"`
	// Diagnostics records parse, type-check, import and limit findings.
	Diagnostics []Diagnostic `json:"diagnostics"`
	// Completeness reports observed coverage without claiming unsupported analysis.
	Completeness Completeness `json:"completeness"`
}
