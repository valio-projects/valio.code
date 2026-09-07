// Package graph builds bounded local Go code graphs.
package graph

// Source is a policy-filtered Go source input. ID must be stable within the
// caller's source snapshot; Content is never copied into Graph output.
type Source struct {
	// ID identifies this local source file independently of any graph version.
	ID string
	// RepositoryID scopes the source to one repository.
	RepositoryID string
	// Path is a slash-separated repository-relative Go source path.
	Path string
	// Content is UTF-8 Go source supplied by the caller.
	Content string
	// ProjectIDs identifies project contexts containing this source.
	ProjectIDs []string
}
