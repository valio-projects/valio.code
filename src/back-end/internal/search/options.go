package search

// Options applies request-local scope, matching, paging, and scan budgets.
type Options struct {
	// ProjectIDs limits results to files with one of these memberships.
	ProjectIDs []string `json:"projectIds,omitempty"`
	// RepositoryIDs limits results to these repositories.
	RepositoryIDs []string `json:"repositoryIds,omitempty"`
	// SnapshotIDs limits results to these immutable snapshots.
	SnapshotIDs []string `json:"snapshotIds,omitempty"`
	// Mode selects term matching; empty selects substring matching.
	Mode Mode `json:"mode,omitempty"`
	// CaseSensitive preserves case rather than Unicode simple-folding it.
	CaseSensitive bool `json:"caseSensitive"`
	// Limit is the maximum returned page size; zero selects the default.
	Limit int `json:"limit,omitempty"`
	// Offset skips that many matched files before paging.
	Offset int `json:"offset,omitempty"`
	// Scan budgets apply before verification. Zero selects bounded defaults of
	// 100,000 files and 256 MiB; exhausting a budget makes Total inexact.
	MaxScanFiles int   `json:"maxScanFiles,omitempty"`
	MaxScanBytes int64 `json:"maxScanBytes,omitempty"`
}
