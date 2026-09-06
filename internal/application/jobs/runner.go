package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// Worker composes leased execution with explicitly registered processors. A
// repository controls durability; processors control bounded application work.
type Worker struct {
	Repository        JobRepository
	Owner             string
	Processors        map[string]Processor
	Concurrency       int
	Logger            *slog.Logger
	PollInterval      time.Duration
	HeartbeatInterval time.Duration
}

func (w *Worker) Run(ctx context.Context) error {
	if w.Repository == nil || w.Owner == "" {
		return errors.New("worker requires repository and owner")
	}
	concurrency := w.Concurrency
	if concurrency == 0 {
		concurrency = 2
	}
	if concurrency < 1 || concurrency > 32 {
		return errors.New("worker concurrency must be 1..32")
	}
	if w.Logger == nil {
		w.Logger = slog.Default()
	}
	if w.PollInterval == 0 {
		w.PollInterval = time.Second
	}
	if w.HeartbeatInterval == 0 {
		w.HeartbeatInterval = 15 * time.Second
	}
	if w.PollInterval < 0 || w.HeartbeatInterval < 0 || w.HeartbeatInterval >= 60*time.Second {
		return errors.New("invalid worker timing")
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, concurrency)
	for i := 0; i < concurrency; i++ {
		go func(slot int) { done <- w.runSlot(child, fmt.Sprintf("%s-%d", w.Owner, slot)) }(i)
	}
	var first error
	for i := 0; i < concurrency; i++ {
		err := <-done
		if first == nil && err != nil {
			first = err
			cancel()
		}
	}
	return first
}
