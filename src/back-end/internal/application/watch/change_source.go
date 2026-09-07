// Package watch reconciles snapshots after filesystem notifications or periodic fallback.
package watch

// ChangeSource supplies coalescible invalidation signals. An event is a reason to
// reconcile authoritative Git/filesystem state, not a complete change journal.
type ChangeSource interface {
	// Events returns coalescible invalidation hints.
	Events() <-chan struct{}
	// Errors reports watcher gaps that require reconciliation.
	Errors() <-chan error
	// Close releases source resources.
	Close() error
}
