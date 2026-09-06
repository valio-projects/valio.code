package search

type Options struct {
	ProjectIDs    []string `json:"projectIds,omitempty"`
	RepositoryIDs []string `json:"repositoryIds,omitempty"`
	SnapshotIDs   []string `json:"snapshotIds,omitempty"`
	Mode          Mode     `json:"mode,omitempty"`
	CaseSensitive bool     `json:"caseSensitive"`
	Limit         int      `json:"limit,omitempty"`
	Offset        int      `json:"offset,omitempty"`
	// Scan budgets apply before verification. Zero selects bounded defaults of
	// 100,000 files and 256 MiB; exhausting a budget makes Total inexact.
	MaxScanFiles int   `json:"maxScanFiles,omitempty"`
	MaxScanBytes int64 `json:"maxScanBytes,omitempty"`
}
