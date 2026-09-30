package handler

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/selis18/go_shortener_url/internal/repository"
)

type recordingDeleteStorage struct {
	batches chan []repository.DeleteRequest
}

func (s recordingDeleteStorage) DeleteBatch(ctx context.Context, batch []repository.DeleteRequest) error {
	s.batches <- batch
	return nil
}

func TestDeletePoolFlush(t *testing.T) {
	for _, mode := range []string{"size", "timer", "shutdown"} {
		t.Run(mode, func(t *testing.T) {
			s := recordingDeleteStorage{make(chan []repository.DeleteRequest, 4)}
			interval := time.Hour
			if mode == "timer" {
				interval = time.Millisecond
			}
			p := newDeletePool(s, 2, 2, interval)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			t.Cleanup(func() { _ = p.shutdown(ctx) })
			tasks := []repository.DeleteRequest{{UserID: "alice", ShortURL: "a"}}
			if mode == "size" {
				tasks = append(tasks, repository.DeleteRequest{UserID: "bob", ShortURL: "b"})
			}
			if err := p.submit(ctx, tasks); err != nil {
				t.Fatal(err)
			}
			if mode == "shutdown" {
				if err := p.shutdown(ctx); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case batch := <-s.batches:
				if len(batch) != len(tasks) {
					t.Fatalf("batch size = %d", len(batch))
				}
				for i := range tasks {
					if batch[i] != tasks[i] {
						t.Fatalf("task changed: %+v", batch[i])
					}
				}
			case <-ctx.Done():
				t.Fatal("batch was not flushed")
			}
		})
	}
}

type waitingDeleteStorage struct{ entered chan struct{} }

func TestDeletePoolFlushesLargeBurstBeforeTimer(t *testing.T) {
	const batchSize = 1000
	const total = 50001
	s := recordingDeleteStorage{make(chan []repository.DeleteRequest, 51)}
	p := newDeletePool(s, 4, batchSize, time.Hour)
	defer p.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tasks := make([]repository.DeleteRequest, total)
	for i := range tasks {
		tasks[i] = repository.DeleteRequest{UserID: "alice", ShortURL: strconv.Itoa(i)}
	}

	if err := p.submit(ctx, tasks[:501]); err != nil {
		t.Fatal(err)
	}
	if err := p.submit(ctx, tasks[501:]); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool, total)
	checkBatch := func(batch []repository.DeleteRequest, wantSize int) {
		t.Helper()
		if len(batch) != wantSize {
			t.Fatalf("batch size = %d, want %d", len(batch), wantSize)
		}
		for _, task := range batch {
			if task.UserID != "alice" || seen[task.ShortURL] {
				t.Fatalf("unexpected or duplicate task: %+v", task)
			}
			seen[task.ShortURL] = true
		}
	}
	for i := 0; i < total/batchSize; i++ {
		select {
		case batch := <-s.batches:
			checkBatch(batch, batchSize)
		case <-ctx.Done():
			t.Fatal("full batches waited for timer")
		}
	}
	if err := p.shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case batch := <-s.batches:
		checkBatch(batch, 1)
	default:
		t.Fatal("shutdown did not flush remainder")
	}
	for _, task := range tasks {
		if !seen[task.ShortURL] {
			t.Fatalf("missing task: %+v", task)
		}
	}
}

func (s waitingDeleteStorage) DeleteBatch(ctx context.Context, batch []repository.DeleteRequest) error {
	s.entered <- struct{}{}
	<-ctx.Done()
	return ctx.Err()
}

func TestDeletePoolParallelWorkersAndCancellation(t *testing.T) {
	s := waitingDeleteStorage{make(chan struct{}, 2)}
	p := newDeletePool(s, 2, 1, time.Hour)
	defer p.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := p.submit(ctx, []repository.DeleteRequest{{UserID: "alice", ShortURL: "a"}, {UserID: "bob", ShortURL: "b"}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		select {
		case <-s.entered:
		case <-ctx.Done():
			t.Fatal("workers did not run concurrently")
		}
	}
	stopCtx, stop := context.WithCancel(context.Background())
	stop()
	_ = p.shutdown(stopCtx)
	select {
	case <-p.done:
	case <-ctx.Done():
		t.Fatal("workers did not stop")
	}
	if err := p.submit(context.Background(), nil); err == nil {
		t.Fatal("accepted tasks after shutdown")
	}
}
