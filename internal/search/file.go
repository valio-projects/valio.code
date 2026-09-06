package search

// File is one immutable searchable source input.
type File struct {
	// ID identifies the file record within the source.
	ID string `json:"id"`
	// Path is a repository-relative path.
	Path string `json:"path"`
	// Content is the exact text searched by this engine.
	Content string `json:"content"`
	// ProjectIDs lists memberships; membership may overlap.
	ProjectIDs []string `json:"projectIds"`
	// RepositoryID scopes Path.
	RepositoryID string `json:"repositoryId"`
	// SnapshotID pins the immutable source version.
	SnapshotID string `json:"snapshotId"`
	// Language is producer-provided metadata for language filters.
	Language string `json:"language"`
	// Test reports producer classification as a test file.
	Test bool `json:"test"`
	// Generated reports producer classification as generated content.
	Generated bool `json:"generated"`
	// Symbols supplies optional symbol and kind filter metadata.
	Symbols []Symbol `json:"symbols,omitempty"`
}
