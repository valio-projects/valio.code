package search

// Result is a page of matched files with scan-completeness evidence.
type Result struct {
	// Matches is the requested result page.
	Matches []Match `json:"matches"`
	// Total is a lower bound when Complete is false. Paging never changes Total.
	Total int `json:"total"`
	// Complete reports whether every eligible file was scanned.
	Complete bool `json:"complete"`
	// Truncated reports either an exhausted scan budget or omitted match ranges.
	// Complete remains true when only ranges were omitted, so Total stays exact.
	Truncated bool `json:"truncated"`
	// ScannedFiles is the number of eligible files inspected before any scan limit.
	ScannedFiles int `json:"scannedFiles"`
	// ScannedBytes is the total UTF-8 source bytes inspected before any scan limit.
	ScannedBytes int64 `json:"scannedBytes"`
}
