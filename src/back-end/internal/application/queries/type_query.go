package queries

// TypeQuery selects an exact type name within an immutable analysis scope.
type TypeQuery struct {
	// Name selects an exact simple or fully qualified type name.
	Name string `json:"name"`
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
	// ProjectID narrows the query to one project in the selected view.
	ProjectID string `json:"projectId"`
	// BuildProfileID narrows type facts to their recorded analysis profile.
	BuildProfileID string `json:"buildProfileId"`
}
