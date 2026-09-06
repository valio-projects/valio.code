package main

import (
	"context"
	"time"
)

// reconcileWatch deliberately uses polling. Each tick rebuilds a static snapshot;
// the caller suppresses unchanged identities and persists before any upload.
func reconcileWatch(ctx context.Context, interval time.Duration, capture func() error) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if e := capture(); e != nil {
				return e
			}
		}
	}
}
