package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	const requestBody = `{"url":"https://example.com/very/long/url"}`

	storage := repository.NewStorageRepo()
	handlers := handler.NewHandlerStorage(storage)
	router := chi.NewRouter()
	router.Use(gzipMiddleware)
	router.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/shorten", handlers.PostShorten)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("decompresses gzip request", func(t *testing.T) {
		var compressedBody bytes.Buffer
		zw := gzip.NewWriter(&compressedBody)
		_, err := zw.Write([]byte(requestBody))
		require.NoError(t, err)
		require.NoError(t, zw.Close())

		req, err := http.NewRequest(http.MethodPost, server.URL+"/api/shorten", &compressedBody)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "identity")

		resp, err := server.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Empty(t, resp.Header.Get("Content-Encoding"))
		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var response model.Response
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response))
		require.True(t, strings.HasPrefix(response.ShortURL, config.GetFlagHost()))
		require.NotEmpty(t, strings.TrimPrefix(response.ShortURL, config.GetFlagHost()))
	})

	t.Run("compresses conflict response for repeated URL", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			server.URL+"/api/shorten",
			bytes.NewBufferString(requestBody),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := server.Client().Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusConflict, resp.StatusCode)
		require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer zr.Close()

		var response model.Response
		require.NoError(t, json.NewDecoder(zr).Decode(&response))
		require.True(t, strings.HasPrefix(response.ShortURL, config.GetFlagHost()))
		require.NotEmpty(t, strings.TrimPrefix(response.ShortURL, config.GetFlagHost()))
	})
}
