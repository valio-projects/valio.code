package structuregraph

// SchemaVersion identifies this syntax-only structure graph encoding.
const SchemaVersion = "structuregraph/syntax/v1"

// Graph is a deterministic, bounded syntactic graph. IDs contain no view ID and
// remain stable only while caller source IDs and helper ranges remain stable.
type Graph struct {
	// Schema identifies the graph contract.
	Schema string `json:"schema"`
	// Language is the syntax helper's normalized language label.
	Language string `json:"language"`
	// ValidSyntax reports the helper's parse result without suppressing partial facts.
	ValidSyntax bool `json:"validSyntax"`
	// Nodes are sorted by ID.
	Nodes []Node `json:"nodes"`
	// Edges are sorted by ID.
	Edges []Edge `json:"edges"`
}
