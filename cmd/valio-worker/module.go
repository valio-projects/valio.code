package main

import (
	"context"
	"fmt"
	"os"

	"github.com/valio-projects/valio.code/internal/application/jobs"
	"github.com/valio-projects/valio.code/internal/configuration"
	"github.com/valio-projects/valio.code/internal/infrastructure/queue"
	"github.com/valio-projects/valio.code/internal/infrastructure/storage/surreal"
	"go.uber.org/fx"
)

var workerModule = fx.Module("worker", fx.Provide(configuration.Database, surreal.New, newWorker), fx.Invoke(runWorker))

func newWorker(db *surreal.Client) *jobs.Worker {
	host, _ := os.Hostname()
	return &jobs.Worker{Repository: queue.Queue{DB: db}, Owner: fmt.Sprintf("%s-%d", host, os.Getpid()), Concurrency: 2,
		Processors: map[string]jobs.Processor{"health": jobs.ProcessorFunc(func(ctx context.Context, _ jobs.Job) error { return db.Ping(ctx) })}}
}

// runWorker binds cancellation and draining to Fx's ordered lifecycle hooks.
func runWorker(lifecycle fx.Lifecycle, shutdown fx.Shutdowner, worker *jobs.Worker, db *surreal.Client) {
	var cancel context.CancelFunc
	done := make(chan struct{})
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.Ping(ctx); err != nil {
				return err
			}
			runCtx, stop := context.WithCancel(context.Background())
			cancel = stop
			go func() {
				defer close(done)
				if err := worker.Run(runCtx); err != nil && runCtx.Err() == nil {
					_ = shutdown.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel == nil {
				return nil
			}
			cancel()
			select {
			case <-done:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	})
}
