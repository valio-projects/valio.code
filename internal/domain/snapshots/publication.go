package snapshots

import "github.com/valio-projects/valio.code/internal/projects"

// Publication is the atomic persistence unit. ExpectedHead fences concurrent
// ingestion; immutable conflicts roll back all staged records and pointer writes.
type Publication struct {
	// ExpectedHead fences publication against concurrent changes to the latest view.
	ExpectedHead string
	// Snapshot carries validated agent input or separately stored snapshot metadata.
	Snapshot Metadata
	// View pins all source inputs, definitions and memberships being atomically published.
	View View
	// Blobs contains immutable, workspace-scoped source content records.
	Blobs []Blob
	// Artifacts contains the complete syntax evidence staged before publication.
	Artifacts []Artifact
	// Manifest contains content-addressed entries and all 256 manifest shard hashes.
	Manifest projects.Manifest
}
