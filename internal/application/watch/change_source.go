// Package watch reconciles snapshots after filesystem notifications or periodic fallback.
package watch

// ChangeSource supplies coalescible invalidation signals. An event is a reason to
// reconcile authoritative Git/filesystem state, not a complete change journal.
type ChangeSource interface {
	Events() <-chan struct{}
	Errors() <-chan error
	Close() error
}
