package scheduling

import (
	"context"
	"time"
)

type Handler func(context.Context, Job) error

func (q Queue) Run(ctx context.Context, owner string, handlers map[string]Handler) error {
	timer := time.NewTicker(time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			j, err := q.Claim(ctx, owner)
			if err != nil {
				continue
			}
			if j == nil {
				continue
			}
			handler, ok := handlers[j.Kind]
			if !ok {
				_ = q.Fail(ctx, *j, "UNSUPPORTED_JOB", false)
				continue
			}
			workCtx, cancel := context.WithCancel(ctx)
			done := make(chan struct{})
			heartbeatDone := make(chan struct{})
			go func() {
				defer close(heartbeatDone)
				t := time.NewTicker(15 * time.Second)
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
			err = handler(workCtx, *j)
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
