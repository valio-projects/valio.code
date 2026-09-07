package repositories

import (
	"context"
	"github.com/valio-projects/valio.code/internal/domain/jobs"
)

// JobRepository is the worker's durable boundary. Queue implements it with
// SurrealDB transactions; tests can exercise processor cancellation independently.
type JobRepository interface {
	Claim(context.Context, string) (*jobs.Job, error)
	Heartbeat(context.Context, jobs.Job) error
	Complete(context.Context, jobs.Job) error
	Fail(context.Context, jobs.Job, string, bool) error
}
