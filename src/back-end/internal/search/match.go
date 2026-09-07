package search

// Match identifies one matched file and its verified source occurrences.
type Match struct {
	// FileID identifies the source file record.
	FileID string `json:"fileId"`
	// Path is the repository-relative matching path.
	Path string `json:"path"`
	// RepositoryID scopes Path.
	RepositoryID string `json:"repositoryId"`
	// SnapshotID pins the source version.
	SnapshotID string `json:"snapshotId"`
	// ProjectIDs are memberships remaining after scope filtering.
	ProjectIDs []string `json:"projectIds"`
	// Ranges are canonical half-open UTF-8 byte occurrences.
	Ranges []Range `json:"ranges"`
	// RangesTruncated reports that matching source occurrences exceeded the per-file limit.
	RangesTruncated bool `json:"rangesTruncated"`
}
