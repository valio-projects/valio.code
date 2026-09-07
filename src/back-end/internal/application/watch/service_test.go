package watch

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeSource struct {
	events chan struct{}
	errors chan error
}

func (f *fakeSource) Events() <-chan struct{} { return f.events }
func (f *fakeSource) Errors() <-chan error    { return f.errors }
func (f *fakeSource) Close() error            { return nil }

func TestWatcherSurvivesFailedDeliveryAndReconciles(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	source := &fakeSource{make(chan struct{}, 10), make(chan error, 1)}
	calls := 0
	failures := 0
	service := Service{Source: source, Interval: 10 * time.Millisecond, Debounce: time.Millisecond, Reconcile: func() error {
		calls++
		if calls == 1 {
			return errors.New("temporary upload failure")
		}
		cancel()
		return nil
	}, OnError: func(error) { failures++ }}
	source.events <- struct{}{}
	source.events <- struct{}{}
	if err := service.Run(ctx); err != nil {
		t.Fatal(err)
	}
	if calls < 2 || failures != 1 {
		t.Fatalf("failed delivery was not retried: %d calls, %d errors", calls, failures)
	}
}
