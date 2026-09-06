package types

// ByNameQuery requests exact case-sensitive simple or qualified type-name matching.
type ByNameQuery struct {
	// Scope supplies required workspace and optional project, profile, and version filters.
	Scope QueryScope `json:"scope"`
	// Name is the exact requested type name.
	Name string `json:"name"`
}
