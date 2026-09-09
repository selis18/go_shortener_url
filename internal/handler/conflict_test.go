package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func TestConcurrentShorteningConflict(t *testing.T) {
	h := NewHandlerStorage(repository.NewStorageRepo())
	type result struct {
		status int
		url    string
	}
	results := make(chan result, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w := httptest.NewRecorder()
			var short string
			if i%2 == 0 {
				h.PostURL(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://same.test")))
				short = w.Body.String()
				if w.Header().Get("Content-Type") != "text/plain" {
					t.Error(w.Header())
				}
			} else {
				h.PostShorten(w, httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(`{"url":"https://same.test"}`)))
				var body model.Response
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Error(err)
				}
				short = body.ShortURL
				if w.Header().Get("Content-Type") != "application/json" {
					t.Error(w.Header())
				}
			}
			results <- result{w.Code, short}
		}(i)
	}
	wg.Wait()
	close(results)
	created := 0
	var expected string
	for result := range results {
		if result.status == http.StatusCreated {
			created++
		} else if result.status != http.StatusConflict {
			t.Errorf("status: %d", result.status)
		}
		if result.url == "" {
			t.Fatal("empty response")
		}
		if expected == "" {
			expected = result.url
		}
		if result.url != expected {
			t.Errorf("different URL: %q, expected %q", result.url, expected)
		}
	}
	if created != 1 {
		t.Fatalf("created %d records, want 1", created)
	}
}
