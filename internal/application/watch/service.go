package watch

import (
	"context"
	"errors"
	"time"
)

// Service batches noisy saves and survives transient capture/upload failures.
type Service struct {
	// Source may be nil; periodic reconciliation still guarantees eventual discovery.
	Source ChangeSource
	// Reconcile must capture and spool before attempting outbound delivery.
	Reconcile func() error
	// OnError receives sanitized diagnostics; it must not log source payloads.
	OnError func(error)
	// Interval is the maximum fallback delay; zero means five minutes.
	Interval time.Duration
	// Debounce and MaxBatch bound event accumulation; zero selects 500ms and 2s.
	Debounce time.Duration
	MaxBatch time.Duration
}

// Run continues until ctx is cancelled. Notifications are hints and are always
// reconciled against a fresh snapshot, including after watcher queue errors.
func (s Service) Run(ctx context.Context) error {
	if s.Reconcile == nil {
		return errors.New("watch service requires a reconciler")
	}
	if s.Interval <= 0 {
		s.Interval = 5 * time.Minute
	}
	if s.Debounce <= 0 {
		s.Debounce = 500 * time.Millisecond
	}
	if s.MaxBatch <= 0 {
		s.MaxBatch = 2 * time.Second
	}
	var events <-chan struct{}
	var failures <-chan error
	if s.Source != nil {
		events = s.Source.Events()
		failures = s.Source.Errors()
		defer s.Source.Close()
	}
	fallback := time.NewTicker(s.Interval)
	defer fallback.Stop()
	timer := time.NewTimer(time.Hour)
	timer.Stop()
	defer timer.Stop()
	var pending <-chan time.Time
	var first time.Time
	reconcile := func() {
		first = time.Time{}
		pending = nil
		if err := s.Reconcile(); err != nil && s.OnError != nil {
			s.OnError(err)
		}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-events:
			if !ok {
				events = nil
				continue
			}
			now := time.Now()
			if first.IsZero() {
				first = now
			}
			delay := min(s.Debounce, s.MaxBatch-now.Sub(first))
			if delay < 0 {
				delay = 0
			}
			timer.Reset(delay)
			pending = timer.C
		case err, ok := <-failures:
			if !ok {
				failures = nil
				continue
			}
			if s.OnError != nil {
				s.OnError(err)
			}
			reconcile()
		case <-pending:
			reconcile()
		case <-fallback.C:
			reconcile()
		}
	}
}
