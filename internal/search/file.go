package search

type File struct {
	ID           string   `json:"id"`
	Path         string   `json:"path"`
	Content      string   `json:"content"`
	ProjectIDs   []string `json:"projectIds"`
	RepositoryID string   `json:"repositoryId"`
	SnapshotID   string   `json:"snapshotId"`
	Language     string   `json:"language"`
	Test         bool     `json:"test"`
	Generated    bool     `json:"generated"`
	Symbols      []Symbol `json:"symbols,omitempty"`
}
