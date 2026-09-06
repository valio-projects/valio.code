package search

type Match struct {
	FileID       string   `json:"fileId"`
	Path         string   `json:"path"`
	RepositoryID string   `json:"repositoryId"`
	SnapshotID   string   `json:"snapshotId"`
	ProjectIDs   []string `json:"projectIds"`
	Ranges       []Range  `json:"ranges"`
}
