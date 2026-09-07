package structuregraph

// Source is one policy-filtered source input. ID must remain stable within the
// caller's immutable source view; Content is used only to validate byte ranges
// and is never copied into a Graph.
type Source struct {
	// ID identifies this source file independently of a graph version.
	ID string
	// RepositoryID scopes this source to one repository in the caller's workspace.
	RepositoryID string
	// Path is the repository-relative path supplied to the syntax helper.
	Path string
	// Content is UTF-8 source used to validate helper-reported byte ranges.
	Content string
	// ProjectIDs records the project contexts that include this source file.
	ProjectIDs []string
}
