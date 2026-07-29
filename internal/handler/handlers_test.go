package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantPost struct {
	code        int
	contentType string
	url         string
}

func Test_PostUrl(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		url  string
		want wantPost
	}{
		{
			name: "simple test",
			url:  "https://practicum.yandex.ru/",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "one symbol",
			url:  "a",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "short url",
			url:  "https://a",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "max long url",
			url:  "https://a" + strings.Repeat("a", 1999),
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "russian alphabet",
			url:  "https://яндекс.ру/",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "special symbols",
			url:  "https://practicum.yandex.ru/?q=hello&lang=ru",
			want: wantPost{
				code:        http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "blank url",
			url:  "",
			want: wantPost{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "only space",
			url:  " ",
			want: wantPost{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Use(middleware.AllowContentType("text/plain"))

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.url))
			request.Header.Set("Content-Type", "text/plain")
			w := httptest.NewRecorder()
			PostUrl(w, request)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			shortUrl := strings.TrimSpace(string(body))

			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.code == http.StatusCreated {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
				assert.NotEmpty(t, shortUrl)
				assert.True(t, strings.HasPrefix(shortUrl, config.FlagHost))

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
		repository.StorageR = repository.NewStorageRepo()
		t.Run(test.name, func(t *testing.T) {

			for k, v := range test.exist {
				repository.StorageR.Storage[k] = v
			}
			path := "/" + test.id

			r := chi.NewRouter()
			r.Get("/{id}", GetUrl)
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
