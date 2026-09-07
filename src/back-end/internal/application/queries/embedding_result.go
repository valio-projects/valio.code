package queries

// EmbeddingResult describes actual indexed coverage and resumable page progress.
type EmbeddingResult struct {
	ViewID             string `json:"viewId"`
	ProfileFingerprint string `json:"profileFingerprint"`
	Eligible           int    `json:"eligible"`
	Processed          int    `json:"processed"`
	NextOffset         int    `json:"nextOffset"`
	Complete           bool   `json:"complete"`
}
