package retrieval

// OmissionReason explains why a relationship candidate was not included.
type OmissionReason string

const (
	// OmissionNotInCaller reports a root ID absent from caller-selected chunks.
	OmissionNotInCaller OmissionReason = "not_in_caller"
	// OmissionProjectScope reports a chunk outside the root or requested project scope.
	OmissionProjectScope OmissionReason = "project_scope"
	// OmissionBudget reports a chunk that would exceed the exact byte budget.
	OmissionBudget OmissionReason = "budget"
	// OmissionDuplicateRange reports a chunk whose canonical range overlaps selected text.
	OmissionDuplicateRange OmissionReason = "duplicate_range"
)

// ContextOmission identifies an omitted relationship candidate and the reason.
type ContextOmission struct {
	// ChunkID identifies the omitted candidate.
	ChunkID string `json:"chunkId"`
	// Reason describes why the candidate was omitted.
	Reason OmissionReason `json:"reason"`
}
