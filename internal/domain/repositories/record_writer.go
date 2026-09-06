package repositories

import "context"

// RecordWriter creates immutable records or saves records explicitly designated mutable.
type RecordWriter[T any] interface {
	// Create rejects an existing key; callers must resolve retries by comparing content.
	Create(ctx context.Context, key string, value T) error
	// Save is forbidden for immutable snapshots, revisions and views.
	Save(ctx context.Context, key string, value T) error
}
