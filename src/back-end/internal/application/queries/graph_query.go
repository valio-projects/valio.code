package queries

import "github.com/valio-projects/valio.code/internal/domain/codegraph"

// GraphMode selects the graph operation applied after immutable view and project
// scope verification.
type GraphMode string

const (
	// GraphSymbols finds symbols with an exact written name.
	GraphSymbols GraphMode = "symbols"
	// GraphReferences finds references whose resolved target is TargetNodeID.
	GraphReferences GraphMode = "references"
	// GraphCallers finds call sites whose resolved target is TargetNodeID.
	GraphCallers GraphMode = "callers"
	// GraphCallees follows calls from a callable symbol or call-site TargetNodeID to their targets.
	GraphCallees GraphMode = "callees"
	// GraphNeighbors traverses incident relations around TargetNodeID.
	GraphNeighbors GraphMode = "neighbors"
	// GraphPaths finds directed paths from SourceNodeID to TargetNodeID.
	GraphPaths GraphMode = "paths"
	// GraphReads finds read occurrences whose resolved target is TargetNodeID.
	GraphReads GraphMode = "reads"
	// GraphWrites finds assignment or initialization occurrences whose resolved target is TargetNodeID.
	GraphWrites GraphMode = "writes"
)

// GraphQuery requests a bounded evidence graph operation in one immutable view.
// EdgeKinds narrows traversal to declared relation kinds; an empty list accepts
// every stored relation kind permitted by the selected mode.
type GraphQuery struct {
	// Mode is required by the HTTP graph endpoint; named MCP tools supply it.
	Mode GraphMode `json:"mode,omitempty"`
	// Scope pins the workspace, immutable view, and optional project memberships.
	Scope SearchScope `json:"scope"`
	// Name is required by GraphSymbols and uses exact Go identifier matching.
	Name string `json:"name,omitempty"`
	// SourceNodeID is required only by GraphPaths and identifies its directed start.
	SourceNodeID string `json:"sourceNodeId,omitempty"`
	// TargetNodeID identifies direct-relation and neighbor targets, or the path destination.
	TargetNodeID string `json:"targetNodeId,omitempty"`
	// EdgeKinds optionally limits returned and traversed relation kinds.
	EdgeKinds []codegraph.RelationKind `json:"edgeKinds,omitempty"`
	// Depth limits neighbor and path traversal. Zero selects the mode default.
	Depth int `json:"depth,omitempty"`
	// Limit bounds direct matches and symbol matches. Zero selects DefaultGraphTargets.
	Limit int `json:"limit,omitempty"`
}

// DefaultGraphDepth is the traversal hop count used when a neighbor or path query omits Depth.
const DefaultGraphDepth = 4

// MaxGraphDepth bounds graph traversal hops independently of artifact size.
const MaxGraphDepth = 12

// DefaultGraphTargets is the direct and name-lookup limit used when Limit is omitted.
const DefaultGraphTargets = 100

// MaxGraphTargets bounds direct and name-lookup matches returned in one query.
const MaxGraphTargets = 1000

// MaxGraphNodes bounds materialized nodes in a traversal response.
const MaxGraphNodes = 2000

// MaxGraphEdges bounds materialized edges in a traversal response.
const MaxGraphEdges = 5000

// MaxGraphPaths bounds distinct directed paths returned by GraphPaths.
const MaxGraphPaths = 100
