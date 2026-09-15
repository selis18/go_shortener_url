package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func TestUserURLsCookie(t *testing.T) {
	s := repository.NewStorageRepo()
	if err := s.Save("a", "https://a.test", "alice"); err != nil {
		t.Fatal(err)
	}
	h := cookieMiddleware(http.HandlerFunc(handler.NewHandlerStorage(s).GetUserURLs))
	sign := func(id string) string {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{UserID: id}).SignedString([]byte(config.GetSecretKey()))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	for _, tc := range []struct {
		name, value string
		status      int
		replacement bool
	}{
		{"absent", "", http.StatusNoContent, true},
		{"malformed", "hello", http.StatusNoContent, true},
		{"missing user ID", sign(""), http.StatusUnauthorized, false},
		{"valid", sign("alice"), http.StatusOK, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			if tc.value != "" {
				r.AddCookie(&http.Cookie{Name: "userId", Value: tc.value})
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status %d", w.Code)
			}
			if (len(w.Result().Cookies()) != 0) != tc.replacement {
				t.Fatal("unexpected cookie replacement")
			}
		})
	}
}
