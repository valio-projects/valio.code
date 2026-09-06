package search

type Result struct {
	Matches []Match `json:"matches"`
	// Total is a lower bound when Complete is false. Paging never changes Total.
	Total        int   `json:"total"`
	Complete     bool  `json:"complete"`
	Truncated    bool  `json:"truncated"`
	ScannedFiles int   `json:"scannedFiles"`
	ScannedBytes int64 `json:"scannedBytes"`
}
