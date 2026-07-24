package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wantPost struct {
	code        int
	contentType string
	url         string
}

func Test_postUrl(t *testing.T) {
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
		StorageR = NewStorageRepo()
		t.Run(test.name, func(t *testing.T) {
			data := url.Values{}
			if test.url != "" {
				data.Set("longUrl", test.url)
			}
			request := httptest.NewRequest(http.MethodPost, "/text/plain", strings.NewReader(data.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			postUrl(w, request)

			res := w.Result()
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			shortUrl := strings.TrimSpace(string(body))

			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.code == http.StatusCreated {
				assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
				assert.NotEmpty(t, shortUrl)
				assert.True(t, strings.HasPrefix(shortUrl, "http://localhost:8080/"))

			}
		})
	}
}

type wantGet struct {
	code     int
	location string
}

func Test_getUrl(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id   string
		want wantGet
	}{
		{
			name: "simple test",
			id:   "xxxxx231",
			want: wantGet{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/",
			},
		},
		{
			name: "exist with long url",
			id:   "asdqw123",
			want: wantGet{
				code:     http.StatusTemporaryRedirect,
				location: "https://practicum.yandex.ru/" + strings.Repeat("a", 1999),
			},
		},
		{
			name: "exist with russian alphabet",
			id:   "a1231asd",
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
	}

	StorageR = NewStorageRepo()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.id != "" && test.want.code == http.StatusTemporaryRedirect {
				StorageR.storage[test.id] = userUrl{
					inputUrl: test.want.location,
					shortUrl: test.id,
				}
			}
			path := "/" + test.id

			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			getUrl(w, request)

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
