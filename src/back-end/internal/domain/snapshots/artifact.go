package snapshots

import (
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/domain/retrieval"
	"github.com/valio-projects/valio.code/internal/domain/typeinfo"
)

// Artifact persists syntax output and scoped descriptors for one view/file.
// Report is encoded JSON so domain persistence contracts do not import parsers.
type Artifact struct {
	// ID identifies this immutable record or registered catalog identity.
	ID string `json:"id"`
	// FileID identifies the source file to which this analysis artifact belongs.
	FileID string `json:"fileId"`
	// ViewID pins the immutable analysis version; an empty query value resolves latest once.
	ViewID string `json:"viewId"`
	// Report stores encoded syntax evidence independently of domain parser implementations.
	Report json.RawMessage `json:"report"`
	// Types contains rich descriptors scoped to the same view, project and syntax profile.
	Types []typeinfo.TypeDescriptor `json:"types"`
	// Chunks are policy-approved source slices with generated metadata kept separate.
	Chunks []retrieval.Chunk `json:"chunks,omitempty"`
	// Graph encodes the versioned codegraph shard owned by this file.
	Graph json.RawMessage `json:"graph,omitempty"`
	// Syntax stores validated non-Go AST facts; unresolved bindings remain explicit.
	Syntax json.RawMessage `json:"syntax,omitempty"`
	// StructureGraph preserves syntax containment separately from compiler-resolved calls.
	StructureGraph json.RawMessage `json:"structureGraph,omitempty"`
}
