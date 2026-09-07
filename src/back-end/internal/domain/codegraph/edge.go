package codegraph

// RelationKind identifies a directed graph relation.
type RelationKind string

const (
	// RelationContains links a package, file, or callable to a syntactic child.
	RelationContains RelationKind = "contains"
	// RelationDeclares links a declaration span to its declared symbol.
	RelationDeclares RelationKind = "declares"
	// RelationRefersTo links an identifier use to a locally resolved symbol.
	RelationRefersTo RelationKind = "refers_to"
	// RelationImports records a written import declaration; its target remains unresolved unless a future provider supplies one.
	RelationImports RelationKind = "imports"
	// RelationMemberOf links a field or method to its locally declared owning type.
	RelationMemberOf RelationKind = "member_of"
	// RelationHasType links a local declaration to its locally declared named type.
	RelationHasType RelationKind = "has_type"
	// RelationCalls links a call site to an exactly resolved local function or method.
	RelationCalls RelationKind = "calls"
	// RelationReads links a source reference to a locally resolved readable variable, constant, or field.
	RelationReads RelationKind = "reads"
	// RelationWrites links an assignment reference, or an initializing definition, to a locally resolved writable target.
	// It is a local access fact and does not represent a data-flow or control-flow edge.
	RelationWrites RelationKind = "writes"
)

// Resolution records how a relation target was established. Syntactic
// containment and declaration relations may be exact from go/parser; semantic
// reference, member, type, and call relations are exact only from go/types.
type Resolution string

const (
	// ResolutionExact means the stated evidence producer established one supplied local target.
	ResolutionExact Resolution = "exact"
	// ResolutionCandidate means candidates were retained without choosing one.
	ResolutionCandidate Resolution = "candidate"
	// ResolutionUnresolved means no target is claimed.
	ResolutionUnresolved Resolution = "unresolved"
)

// Edge is one directed, evidence-carrying relation. TargetID is empty only for
// an unresolved relation; consumers must not interpret that state as no target.
type Edge struct {
	// ID is stable for the relation endpoints, kind, range and resolution.
	ID string `json:"id"`
	// Kind classifies the directed relation.
	Kind RelationKind `json:"kind"`
	// SourceID identifies the originating node.
	SourceID string `json:"sourceId"`
	// TargetID identifies the exact or candidate target when available.
	TargetID string `json:"targetId,omitempty"`
	// Resolution states whether TargetID is exact, candidate, or absent.
	Resolution Resolution `json:"resolution"`
	// ProjectIDs is the compatible project-context intersection for this relation.
	ProjectIDs []string `json:"projectIds"`
	// Evidence identifies the source observation or semantic resolution.
	Evidence Evidence `json:"evidence"`
}
