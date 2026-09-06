package main

import (
	"context"
	"fmt"
	appwatch "github.com/valio-projects/valio.code/internal/application/watch"
	infra "github.com/valio-projects/valio.code/internal/infrastructure/watch"
	"os"
	"time"
)

// reconcileWatch combines fsnotify hints with authoritative periodic capture.
func reconcileWatch(ctx context.Context, root string, interval time.Duration, capture func() error) error {
	source, err := infra.NewFSNotifySource(root)
	onError := func(err error) { fmt.Fprintln(os.Stderr, "watch:", err) }
	if err != nil {
		onError(err)
		return (appwatch.Service{Interval: interval, Reconcile: capture, OnError: onError}).Run(ctx)
	}
	return (appwatch.Service{Source: source, Interval: interval, Reconcile: capture, OnError: onError}).Run(ctx)
}
