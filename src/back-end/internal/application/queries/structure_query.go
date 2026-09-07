package queries

// StructureQuery selects syntax containment in a pinned project view.
// Name matches declarations. NodeID can select any observed graph node.
type StructureQuery struct {
	Scope  SearchScope `json:"scope"`            // Scope fixes workspace, projects and source version.
	Name   string      `json:"name,omitempty"`   // Name is an exact written declaration name.
	FileID string      `json:"fileId,omitempty"` // FileID optionally restricts the source file.
	NodeID string      `json:"nodeId,omitempty"` // NodeID selects a previously returned observation.
	Depth  int         `json:"depth,omitempty"`  // Depth bounds outgoing containment, default 2, maximum 12.
	Limit  int         `json:"limit,omitempty"`  // Limit bounds returned nodes across files, default 200, maximum 2000.
}
