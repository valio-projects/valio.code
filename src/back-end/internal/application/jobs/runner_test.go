package jobs

import (
	"context"
	"sync"
	"testing"
	"time"
)

type testJobRepository struct {
	mu        sync.Mutex
	remaining int
	completed int
	loseLease bool
}

func (r *testJobRepository) Claim(_ context.Context, owner string) (*Job, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.remaining == 0 {
		return nil, nil
	}
	r.remaining--
	return &Job{ID: owner, Kind: "test", Owner: owner, Fence: 1}, nil
}
func (r *testJobRepository) Heartbeat(context.Context, Job) error {
	if r.loseLease {
		return ErrLeaseLost
	}
	return nil
}
func (r *testJobRepository) Complete(context.Context, Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.completed++
	return nil
}
func (r *testJobRepository) Fail(context.Context, Job, string, bool) error { return nil }

func TestWorkerProcessesTwoJobsAndCancels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	entered := make(chan string, 2)
	repo := &testJobRepository{remaining: 2}
	processor := ProcessorFunc(func(ctx context.Context, j Job) error { entered <- j.ID; <-ctx.Done(); return ctx.Err() })
	worker := &Worker{Repository: repo, Owner: "test", Processors: map[string]Processor{"test": processor}, PollInterval: time.Millisecond}
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	for i := 0; i < 2; i++ {
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("jobs were not processed concurrently")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.completed != 0 {
		t.Fatal("cancelled jobs completed")
	}
}
func TestLeaseLossCancelsProcessor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stopped := make(chan struct{})
	worker := &Worker{Repository: &testJobRepository{remaining: 1, loseLease: true}, Owner: "test", Concurrency: 1, PollInterval: time.Millisecond, HeartbeatInterval: time.Millisecond,
		Processors: map[string]Processor{"test": ProcessorFunc(func(ctx context.Context, _ Job) error { <-ctx.Done(); close(stopped); return ctx.Err() })}}
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	select {
	case <-stopped:
	case <-ctx.Done():
		t.Fatal("lost lease did not cancel processor")
	}
	cancel()
	<-done
}
