package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func TestShortenBatchGzip(t *testing.T) {
	storage := repository.NewStorageRepo()
	h := handler.NewHandlerStorage(storage)
	var body bytes.Buffer
	zw := gzip.NewWriter(&body)
	zw.Write([]byte(`[{"correlation_id":"first","original_url":"https://a.test"},{"correlation_id":"second","original_url":"https://a.test"}]`))
	zw.Close()
	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", &body)
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	gzipMiddleware(http.HandlerFunc(h.PostShortenBatch)).ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("status: %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatal(w.Header())
	}
	zr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	var response []model.BatchResponse
	if err := json.NewDecoder(zr).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if len(response) != 2 || response[0].CorrelationID != "first" || response[1].CorrelationID != "second" || response[0].ShortURL != response[1].ShortURL {
		t.Fatalf("response: %+v", response)
	}
	key, ok := storage.FindByValue("https://a.test")
	if !ok || !strings.HasSuffix(response[0].ShortURL, key) {
		t.Fatal("URL not saved")
	}
}

func TestShortenBatchInvalid(t *testing.T) {
	for _, body := range []string{"", "null", "[]", "{}", "[", `[{"original_url":""}]`, `[{"original_url":"https://a.test"}] {}`, `[{"original_url":"https://a.test"},{"original_url":" "}]`} {
		t.Run(body, func(t *testing.T) {
			s := repository.NewStorageRepo()
			w := httptest.NewRecorder()
			handler.NewHandlerStorage(s).PostShortenBatch(w, httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body)))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status: %d", w.Code)
			}
			if _, ok := s.FindByValue("https://a.test"); ok {
				t.Fatal("invalid batch saved")
			}
		})
	}
}
