package jobs

import "context"

type Processor interface {
	Process(context.Context, Job) error
}
type ProcessorFunc func(context.Context, Job) error

func (f ProcessorFunc) Process(ctx context.Context, j Job) error { return f(ctx, j) }
