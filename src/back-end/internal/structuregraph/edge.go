package structuregraph

// RelationKind identifies a syntactic graph relation.
type RelationKind string

const (
	// RelationContains links a source container to an observed child fact.
	RelationContains RelationKind = "contains"
	// RelationDeclares links a file or declaration to a declaration it syntactically owns.
	RelationDeclares RelationKind = "declares"
	// RelationHasParameter links a declaration to one reported parameter.
	RelationHasParameter RelationKind = "has_parameter"
	// RelationImports records a written import with no resolved target claim.
	RelationImports RelationKind = "imports"
)

// Resolution states whether a relation target was syntactically established.
type Resolution string

const (
	// ResolutionExact means both endpoints follow directly from helper structure.
	ResolutionExact Resolution = "exact"
	// ResolutionUnresolved means the observation deliberately has no target node.
	ResolutionUnresolved Resolution = "unresolved"
)

// Edge is one directed, evidence-carrying syntactic relation.
type Edge struct {
	// ID is deterministic from relation kind, endpoints, resolution, and evidence range.
	ID string `json:"id"`
	// Kind classifies the relation.
	Kind RelationKind `json:"kind"`
	// SourceID identifies the owning or observing node.
	SourceID string `json:"sourceId"`
	// TargetID identifies the syntactically known target. It is empty only for unresolved imports.
	TargetID string `json:"targetId,omitempty"`
	// Resolution says whether TargetID is a syntax-established node.
	Resolution Resolution `json:"resolution"`
	// ProjectIDs records the source contexts in which this relation was observed.
	ProjectIDs []string `json:"projectIds"`
	// Evidence identifies the helper observation supporting this relation.
	Evidence Evidence `json:"evidence"`
}
