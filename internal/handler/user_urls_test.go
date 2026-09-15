package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func TestGetUserURLs(t *testing.T) {
	s := repository.NewStorageRepo()
	if err := s.Save("a", "https://alice.test", "alice"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save("b", "https://bob.test", "bob"); err != nil {
		t.Fatal(err)
	}
	h := NewHandlerStorage(s)
	for _, tc := range []struct {
		name, user string
		status     int
		cancelled  bool
	}{
		{"missing ID", "", http.StatusUnauthorized, false},
		{"empty list", "new", http.StatusNoContent, false},
		{"own URLs", "alice", http.StatusOK, false},
		{"storage error", "alice", http.StatusInternalServerError, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			ctx := auth.WithUserID(r.Context(), tc.user)
			if tc.cancelled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			w := httptest.NewRecorder()
			h.GetUserURLs(w, r.WithContext(ctx))
			if w.Code != tc.status {
				t.Fatalf("status %d", w.Code)
			}
			if tc.status != http.StatusOK {
				if w.Body.Len() != 0 {
					t.Fatal("expected empty body")
				}
				return
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Fatal("content type")
			}
			var response []map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if len(response) != 1 || len(response[0]) != 2 || response[0]["short_url"] != config.GetFlagHost()+"a" || response[0]["original_url"] != "https://alice.test" {
				t.Fatalf("response: %s", w.Body.String())
			}
		})
	}
}
