package queries

import "github.com/valio-projects/valio.code/internal/domain/codegraph"

// GraphStatus reports whether the selected immutable view supplied complete,
// partial, or unsupported graph evidence for a query.
type GraphStatus string

const (
	// GraphComplete means every selected graph shard was decoded and no response limit stopped traversal.
	GraphComplete GraphStatus = "complete"
	// GraphPartial means graph publication or the response traversal was incomplete.
	GraphPartial GraphStatus = "partial"
	// GraphUnsupported means the selected view predates, or lacks, this graph projection.
	GraphUnsupported GraphStatus = "unsupported"
)

// GraphNoticeCode classifies an explicit limitation without turning valid
// evidence into a stronger claim.
type GraphNoticeCode string

const (
	// GraphUnsupportedOldView identifies a view without a compatible graph shard.
	GraphUnsupportedOldView GraphNoticeCode = "unsupported_old_view"
	// GraphMissingShard identifies a selected Go artifact without graph evidence.
	GraphMissingShard GraphNoticeCode = "missing_graph_shard"
	// GraphNodeBudget identifies a response stopped by node, edge, target, or path bounds.
	GraphNodeBudget GraphNoticeCode = "node_budget"
	// GraphCandidateTarget identifies a returned relation that has candidates rather than an exact target.
	GraphCandidateTarget GraphNoticeCode = "candidate_target"
)

// GraphNotice describes an evidence or response limitation in a machine-readable form.
type GraphNotice struct {
	// Code identifies the limitation category.
	Code GraphNoticeCode `json:"code"`
	// Message describes the limitation without copying source content.
	Message string `json:"message"`
}

// GraphResult returns graph facts and their original parser/type-checker evidence.
// Paths contain ordered node IDs; their constituent relations are included in Edges.
type GraphResult struct {
	// ViewID is the immutable analysis view resolved once for this result.
	ViewID string `json:"viewId"`
	// Status reports graph availability and bounded-response completeness.
	Status GraphStatus `json:"status"`
	// Nodes carries source ranges, project scope, and evidence producers.
	Nodes []codegraph.Node `json:"nodes"`
	// Edges carries typed relations, resolution state, scope, and evidence producers.
	Edges []codegraph.Edge `json:"edges"`
	// Paths contains directed paths for GraphPaths only.
	Paths [][]string `json:"paths,omitempty"`
	// Diagnostics preserves safe parser and type-check findings from accepted shards.
	Diagnostics []codegraph.Diagnostic `json:"diagnostics,omitempty"`
	// Completeness describes the published builder result without claiming compiler coverage.
	Completeness codegraph.Completeness `json:"completeness"`
	// Notices makes unsupported data, response budgets, and candidate targets explicit.
	Notices []GraphNotice `json:"notices,omitempty"`
}
