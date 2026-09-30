package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/repository"
)

type blockingDeleteStorage struct {
	*repository.StorageRepo
	entered chan context.Context
	release chan struct{}
	done    chan struct{}
}

func (s *blockingDeleteStorage) DeleteBatch(ctx context.Context, batch []repository.DeleteRequest) error {
	s.entered <- ctx
	<-s.release
	err := s.StorageRepo.DeleteBatch(ctx, batch)
	close(s.done)
	return err
}

func TestDeleteIsAsyncAndReturnsGone(t *testing.T) {
	s := &blockingDeleteStorage{repository.NewStorageRepo(), make(chan context.Context, 1), make(chan struct{}), make(chan struct{})}
	if err := s.Save("a", "https://a", "alice"); err != nil {
		t.Fatal(err)
	}
	h := NewHandlerStorage(s)
	h.StartDeletionWorkers()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := h.ShutdownDeletes(ctx); err != nil {
			t.Errorf("shutdown deletion workers: %v", err)
		}
	})
	ctx, cancel := context.WithCancel(auth.WithUserID(context.Background(), "alice"))
	defer cancel()
	r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`["a"]`)).WithContext(ctx)
	w := httptest.NewRecorder()
	returned := make(chan struct{})
	go func() { h.DeleteUserURLs(w, r); close(returned) }()
	defer close(s.release)
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("handler waited for storage")
	}
	if w.Code != http.StatusAccepted {
		t.Fatal(w.Code)
	}
	cancel()
	select {
	case workCtx := <-s.entered:
		if workCtx.Err() != nil {
			t.Fatal("request cancellation reached worker")
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}
	s.release <- struct{}{}
	select {
	case <-s.done:
	case <-time.After(time.Second):
		t.Fatal("worker did not finish")
	}
	router := chi.NewRouter()
	router.Get("/{id}", h.GetURL)
	result := httptest.NewRecorder()
	router.ServeHTTP(result, httptest.NewRequest(http.MethodGet, "/a", nil))
	if result.Code != http.StatusGone || result.Header().Get("Location") != "" {
		t.Fatalf("response: %v", result)
	}
}

func TestDeleteValidation(t *testing.T) {
	h := NewHandlerStorage(repository.NewStorageRepo())
	for _, body := range []string{`null`, `{}`, `[1]`, `[""]`, `["a"] []`, `["a"`, `[null]`} {
		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(body))
		r = r.WithContext(auth.WithUserID(r.Context(), "alice"))
		w := httptest.NewRecorder()
		h.DeleteUserURLs(w, r)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d", body, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.DeleteUserURLs(w, httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(`[]`)))
	if w.Code != http.StatusUnauthorized {
		t.Fatal(w.Code)
	}
}
