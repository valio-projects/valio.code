package jobs

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"time"
)

func (w *Worker) runSlot(ctx context.Context, owner string) error {
	q := w.Repository
	timer := time.NewTicker(w.PollInterval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			j, err := q.Claim(ctx, owner)
			if err != nil {
				w.Logger.Warn("job claim failed", "worker", owner, "error_class", "database")
				continue
			}
			if j == nil {
				continue
			}
			processor, ok := w.Processors[j.Kind]
			if !ok {
				_ = q.Fail(ctx, *j, "UNSUPPORTED_JOB", false)
				continue
			}
			workCtx, cancel := context.WithCancel(ctx)
			workCtx = propagation.TraceContext{}.Extract(workCtx, propagation.MapCarrier{"traceparent": j.TraceParent})
			workCtx, span := otel.Tracer("valio.worker").Start(workCtx, "job.process")
			span.SetAttributes(attribute.String("job.id", j.ID), attribute.String("job.kind", j.Kind), attribute.Int("job.fence", j.Fence))
			done := make(chan struct{})
			heartbeatDone := make(chan struct{})
			go func() {
				defer close(heartbeatDone)
				t := time.NewTicker(w.HeartbeatInterval)
				defer t.Stop()
				for {
					select {
					case <-done:
						return
					case <-workCtx.Done():
						return
					case <-t.C:
						if q.Heartbeat(workCtx, *j) != nil {
							cancel()
							return
						}
					}
				}
			}()
			err = processor.Process(workCtx, *j)
			if err != nil {
				span.SetStatus(codes.Error, "processor_failure")
			}
			span.End()
			close(done)
			cancel()
			<-heartbeatDone
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				_ = q.Fail(ctx, *j, "HANDLER_FAILED", true)
			} else {
				_ = q.Complete(ctx, *j)
			}
		}
	}
}
