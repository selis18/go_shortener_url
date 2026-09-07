package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type pingerFunc func(context.Context) error

func (f pingerFunc) PingContext(ctx context.Context) error { return f(ctx) }

func TestPingHandler(t *testing.T) {
	for _, tc := range []struct {
		name string
		db   DatabasePinger
		want int
	}{
		{"connected", pingerFunc(func(context.Context) error { return nil }), http.StatusOK},
		{"unavailable", pingerFunc(func(context.Context) error { return errors.New("unavailable") }), http.StatusInternalServerError},
		{"not configured", nil, http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			NewPingHandler(tc.db)(res, httptest.NewRequest(http.MethodGet, "/ping", nil))
			if res.Code != tc.want {
				t.Fatalf("status = %d, want %d", res.Code, tc.want)
			}
		})
	}
}

func TestPingHandlerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	db := pingerFunc(func(ctx context.Context) error {
		called = true
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > 5*time.Second {
			t.Error("expected a deadline within 5 seconds")
		}
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Error("request cancellation was not propagated")
		}
		return ctx.Err()
	})
	res := httptest.NewRecorder()
	NewPingHandler(db)(res, httptest.NewRequest(http.MethodGet, "/ping", nil).WithContext(ctx))
	if !called {
		t.Fatal("database was not checked")
	}
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.Code)
	}
}
