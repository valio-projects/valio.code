package repositories

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain/snapshots"
	"github.com/valio-projects/valio.code/internal/search"
)

// SnapshotRepository resolves workspace-bound immutable publications and
// bounded source/artifact reads. Publish must atomically fence ExpectedHead.
type SnapshotRepository interface {
	CatalogRepository
	// Latest resolves the current publication pointer into one immutable view; ctx controls cancellation.
	Latest(context.Context) (snapshots.View, error)
	// View reads id as an immutable view and verifies workspace scope; ctx controls cancellation.
	View(context.Context, string) (snapshots.View, error)
	// Publish atomically stores p after validating its expected head and immutable entries; ctx controls cancellation.
	Publish(context.Context, snapshots.Publication) error
	// Files loads only source references in v and rejects exhausted scope budgets; ctx controls cancellation.
	Files(context.Context, snapshots.View) ([]search.File, error)
	// Artifacts loads syntax/type evidence belonging only to v; ctx controls cancellation.
	Artifacts(context.Context, snapshots.View) ([]snapshots.Artifact, error)
}
