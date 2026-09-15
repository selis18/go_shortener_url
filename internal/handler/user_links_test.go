package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/repository"
)

type userRecordingStorage struct {
	*repository.StorageRepo
	userID string
}

func (s *userRecordingStorage) SaveContext(ctx context.Context, k, v, userID string) error {
	s.userID = userID
	return s.StorageRepo.SaveContext(ctx, k, v, userID)
}

func (s *userRecordingStorage) SaveBatch(ctx context.Context, pairs []repository.URLPair, userID string) ([]string, error) {
	s.userID = userID
	return s.StorageRepo.SaveBatch(ctx, pairs, userID)
}

func TestHandlersPassUserID(t *testing.T) {
	for _, name := range []string{"plain", "json", "batch"} {
		t.Run(name, func(t *testing.T) {
			s := &userRecordingStorage{StorageRepo: repository.NewStorageRepo()}
			h := NewHandlerStorage(s)
			body := "https://example.test"
			handle := h.PostURL
			if name == "json" {
				body = `{"url":"https://example.test"}`
				handle = h.PostShorten
			}
			if name == "batch" {
				body = `[{"correlation_id":"1","original_url":"https://example.test"}]`
				handle = h.PostShortenBatch
			}
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			r = r.WithContext(auth.WithUserId(r.Context(), "alice"))
			w := httptest.NewRecorder()
			handle(w, r)
			if w.Code != http.StatusCreated {
				t.Fatalf("status: %d", w.Code)
			}
			if s.userID != "alice" {
				t.Fatalf("user ID: %q", s.userID)
			}
		})
	}
}
