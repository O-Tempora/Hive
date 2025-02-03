package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/O-Tempora/hive/worker"
)

func TestWorker_Basics(t *testing.T) {
	t.Parallel()

	t.Run("expecting to run at least 5 times in a second with no limitations", func(t *testing.T) {
		t.Parallel()

		tsk, done := taskCounter(5)
		wk := worker.NewBuilder().Task(tsk).Build()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := worker.StartBackgroundWorker(ctx, wk)
		if err != nil {
			t.Errorf("worker error: %s", err.Error())
		}

		select {
		case <-ctx.Done():
			t.Fatal("context Done() had fired before task execution count was reached")
		case <-done:
			// noop
		}
	})

	t.Run("should stop after parent context cancelation", func(t *testing.T) {
		t.Parallel()

		tsk, done := taskCounter(0)
		wk := worker.NewBuilder().Task(tsk).Build()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		cancel()

		err := worker.StartBackgroundWorker(ctx, wk)
		if err != nil {
			t.Errorf("worker error: %s", err.Error())
		}

		select {
		case <-ctx.Done():
			// noop
		case <-done:
			t.Fatal("parent context is canceled therefore task must not be executed")
		}
	})
}

func TestWorker_Validation(t *testing.T) {
	t.Parallel()

	t.Run("worker is nil", func(t *testing.T) {
		t.Parallel()

		err := worker.StartBackgroundWorker(context.Background(), nil)
		if err == nil {
			t.Error("must return an error since worker is nil")
		}
	})

	t.Run("task is nil", func(t *testing.T) {
		t.Parallel()

		err := worker.StartBackgroundWorker(
			context.Background(),
			worker.NewBuilder().Task(nil).Build(),
		)
		if err == nil {
			t.Error("must return an error since task is nil")
		}
	})
}

func taskCounter(count uint32) (worker.Task, <-chan struct{}) {
	var ct atomic.Uint32
	doneCh := make(chan struct{})

	task := worker.Task(func(ctx context.Context) error {
		if ct.Add(1) == count {
			close(doneCh)
		}

		return nil
	})

	return task, doneCh
}
