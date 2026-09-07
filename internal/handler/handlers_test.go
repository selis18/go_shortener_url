package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerStorage_PostShorten(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantURL    string
	}{
		{
			name:       "valid URL",
			body:       `{"url":"https://practicum.yandex.ru/"}`,
			wantStatus: http.StatusCreated,
			wantURL:    "https://practicum.yandex.ru/",
		},
		{
			name:       "URL with query parameters",
			body:       `{"url":"https://practicum.yandex.ru/search?q=go&lang=ru"}`,
			wantStatus: http.StatusCreated,
			wantURL:    "https://practicum.yandex.ru/search?q=go&lang=ru",
		},
		{
			name:       "URL with cyrillic characters",
			body:       `{"url":"https://яндекс.рф/поиск"}`,
			wantStatus: http.StatusCreated,
			wantURL:    "https://яндекс.рф/поиск",
		},
		{
			name:       "additional JSON field",
			body:       `{"url":"https://practicum.yandex.ru/","extra":"ignored"}`,
			wantStatus: http.StatusCreated,
			wantURL:    "https://practicum.yandex.ru/",
		},
		{
			name:       "empty URL",
			body:       `{"url":""}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing URL field",
			body:       `{}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "malformed JSON",
			body:       `{"url":`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrong URL type",
			body:       `{"url":123}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "empty body",
			body:       "",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := repository.NewStorageRepo()
			handler := NewHandlerStorage(storage)
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			responseRecorder := httptest.NewRecorder()

			handler.PostShorten(responseRecorder, request)

			response := responseRecorder.Result()
			defer response.Body.Close()
			require.Equal(t, test.wantStatus, response.StatusCode)

			if test.wantStatus != http.StatusCreated {
				return
			}

			assert.Contains(t, response.Header.Get("Content-Type"), "application/json")
			var responseBody model.Response
			require.NoError(t, json.NewDecoder(response.Body).Decode(&responseBody))
			require.True(t, strings.HasPrefix(responseBody.ShortURL, config.GetFlagHost()))

			shortID := strings.TrimPrefix(responseBody.ShortURL, config.GetFlagHost())
			require.NotEmpty(t, shortID)
			savedURL, err := storage.Get(shortID)
			require.NoError(t, err)
			assert.Equal(t, test.wantURL, savedURL)
		})
	}
}

type wantPost struct {
	code        int
	contentType string
	URL         string
}

func Test_PostUrl(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		URL  string
		want wantPost
	}{
		{
			name: "simple test",
			URL:  "https://practicum.yandex.ru/",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "one symbol",
			URL:  "a",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "short url",
			URL:  "https://a",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "max long url",
			URL:  "https://a" + strings.Repeat("a", 1999),
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "russian alphabet",
			URL:  "https://яндекс.ру/",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "special symbols",
			URL:  "https://practicum.yandex.ru/?q=hello&lang=ru",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "blank url",
			URL:  "",
			want: wantPost{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "only space",
			URL:  " ",
			want: wantPost{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := repository.NewStorageRepo()
			handler := NewHandlerStorage(storage)
			r := chi.NewRouter()
			r.Use(middleware.AllowContentType("text/plain"))
			r.Post("/", handler.PostURL)

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.URL))
			request.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			shortURL := strings.TrimSpace(string(body))

			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.code == http.StatusCreated {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
				assert.NotEmpty(t, shortURL)
				assert.True(t, strings.HasPrefix(shortURL, config.GetFlagHost()))

			}
		})
	}
}

type wantGet struct {
	code     int
	location string
}

func Test_GetUrl(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		exist map[string]string
		id    string
		want  wantGet
	}{
		{
			name: "simple test",
			exist: map[string]string{
				"xxxxx231": "https://practicum.yandex.ru/",
			},
			id: "xxxxx231",
			want: wantGet{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "exist with long url",
			exist: map[string]string{
				"asdqw123": "https://practicum.yandex.ru/" + strings.Repeat("a", 1999),
			},
			id: "asdqw123",
			want: wantGet{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/" + strings.Repeat("a", 1999),
			},
		},
		{
			name: "exist with russian alphabet",
			exist: map[string]string{
				"a1231asd": "https://яндекс.ру/",
			},
			id: "a1231asd",
			want: wantGet{
				code:     http.StatusTemporaryRedirect,
				location: "https://яндекс.ру/",
			},
		},
		{
			name: "without parametrs",
			id:   "",
			want: wantGet{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no result",
			id:   "a1231asd",
			want: wantGet{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		storage := repository.NewStorageRepo()
		handler := NewHandlerStorage(storage)
		t.Run(test.name, func(t *testing.T) {

			for k, v := range test.exist {
				handler.storage.Save(k, v)
			}
			path := "/" + test.id

			r := chi.NewRouter()
			r.Get("/{id}", handler.GetURL)
			r.NotFound(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})

			request := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)

			if test.want.code == http.StatusTemporaryRedirect {
				location := res.Header.Get("Location")
				assert.Equal(t, test.want.location, location)
			}
		})
	}
}
