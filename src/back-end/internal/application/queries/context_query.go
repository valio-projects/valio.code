package queries

// ContextQuery selects bounded parent/child source context within one view.
type ContextQuery struct {
	Scope    SearchScope `json:"scope"`
	ChunkID  string      `json:"chunkId"`
	MaxBytes int         `json:"maxBytes,omitempty"`
}
