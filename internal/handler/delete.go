package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/repository"
	"go.uber.org/zap"
)

func (h *HandlerStorage) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	defer r.Body.Close()
	var ids []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ids); err != nil || ids == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, id := range ids {
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if h.deletes == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tasks := make([]repository.DeleteRequest, len(ids))
	for i, id := range ids {
		tasks[i] = repository.DeleteRequest{UserID: userID, ShortURL: id}
	}
	if err := h.deletes.submit(r.Context(), tasks); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type deletePool struct {
	mu     sync.Mutex
	closed bool
	input  chan []repository.DeleteRequest
	done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

func (h *HandlerStorage) StartDeletionWorkers() {
	if h.deletes != nil {
		return
	}
	if storage, ok := h.storage.(repository.DeleteStorage); ok {
		h.deletes = newDeletePool(storage, 4, config.GetDeleteBatchSize(), config.GetDeleteFlushInterval())
	}
}

func (h *HandlerStorage) ShutdownDeletes(ctx context.Context) error {
	if h.deletes == nil {
		return nil
	}
	return h.deletes.shutdown(ctx)
}

func newDeletePool(storage repository.DeleteStorage, workers, batchSize int, interval time.Duration) *deletePool {
	ctx, cancel := context.WithCancel(context.Background())
	p := &deletePool{input: make(chan []repository.DeleteRequest, 64), done: make(chan struct{}), ctx: ctx, cancel: cancel}
	jobs := make(chan []repository.DeleteRequest, workers)
	// Fan-in: all workers send their results to a single collector.
	results := make(chan error, workers)
	var wg sync.WaitGroup
	wg.Add(workers + 1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		batch := make([]repository.DeleteRequest, 0, batchSize)
		flush := func() bool {
			if len(batch) == 0 {
				return true
			}
			select {
			case jobs <- batch:
				batch = make([]repository.DeleteRequest, 0, batchSize)
				return true
			case <-ctx.Done():
				return false
			}
		}
		for {
			select {
			case <-ctx.Done():
				return
			case tasks, ok := <-p.input:
				if !ok {
					flush()
					return
				}
				for _, task := range tasks {
					batch = append(batch, task)
					if len(batch) == batchSize && !flush() {
						return
					}
				}
			case <-ticker.C:
				if !flush() {
					return
				}
			}
		}
	}()
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-jobs:
					if !ok {
						return
					}
					workCtx, stop := context.WithTimeout(ctx, 5*time.Second)
					err := storage.DeleteBatch(workCtx, batch)
					stop()
					results <- err
				}
			}
		}()
	}
	go func() { wg.Wait(); close(results) }()
	go func() {
		defer close(p.done)
		defer cancel()
		for err := range results {
			if err != nil {
				logger.Log.Error("delete URLs", zap.Error(err))
			}
		}
	}()
	return p
}

func (p *deletePool) submit(ctx context.Context, tasks []repository.DeleteRequest) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return context.Canceled
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.input <- tasks:
		return nil
	}
}

func (p *deletePool) shutdown(ctx context.Context) error {
	stop := context.AfterFunc(ctx, p.cancel)
	defer stop()
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.input)
	}
	p.mu.Unlock()
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		p.cancel()
		return ctx.Err()
	}
}
