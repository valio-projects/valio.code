// Package retrieval builds bounded, source-grounded chunks and ranks them.
package retrieval

// Source supplies one immutable, already policy-filtered source file.
type Source struct {
	// ID identifies the immutable file record.
	ID string
	// RepositoryID scopes the source repository.
	RepositoryID string
	// Path is the repository-relative source path used in generated metadata.
	Path string
	// Language identifies the source language, such as "go".
	Language string
	// Content contains canonical UTF-8 source text.
	Content string
	// ProjectIDs contains project memberships that must remain attached to chunks.
	ProjectIDs []string
}
