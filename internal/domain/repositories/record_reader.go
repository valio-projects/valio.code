package repositories

import "context"

// RecordReader retrieves typed records within an already-authorized workspace.
// Keys are stable identities; pagination must be deterministic and bounded.
type RecordReader[T any] interface {
	// Find returns the record identified by key or the implementation's not-found error.
	Find(ctx context.Context, key string) (T, error)
	// List returns at most limit records after offset in stable identity order.
	List(ctx context.Context, limit, offset int) ([]T, error)
}
