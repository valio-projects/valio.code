package snapshots

// IngestResult returns the immutable snapshot/view receipt and actual completeness.
type IngestResult struct {
	// SnapshotID pins the concrete source snapshot used by this record.
	SnapshotID string `json:"snapshotId"`
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
	// Status reports actual completeness or command outcome without implying compiler evidence.
	Status string `json:"status"`
}
