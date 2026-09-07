package queries

// StructureResult returns bounded syntax relationships, never resolved execution.
type StructureResult struct {
	ViewID      string                `json:"viewId"`      // ViewID pins source and analyzer versions.
	Projection  string                `json:"projection"`  // Projection identifies the structure schema.
	Status      string                `json:"status"`      // Status is partial syntax or unsupported capability.
	Files       []StructureFileResult `json:"files"`       // Files separates independent source graphs.
	Truncated   bool                  `json:"truncated"`   // Truncated reports traversal or output budget exhaustion.
	StopReasons []string              `json:"stopReasons"` // StopReasons explains bounded omissions.
}
