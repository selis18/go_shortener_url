package handler

import (
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/repository"
)

type HandlerStorage struct {
	storage repository.Storage
}

func NewHandlerStorage(s repository.Storage) *HandlerStorage {
	return &HandlerStorage{
		storage: s,
	}
}

const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNOPQRSTUVWXYZ23456789"

func generateID() string {
	id := ""
	for i := 0; i < 10; i++ {
		randInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			panic(err)
		}
		id += string(alphabet[randInt.Int64()])
	}
	return id
}

func (h *HandlerStorage) PostUrl(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	longUrl := strings.TrimSpace(string(body))
	if longUrl == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if exKey, found := h.storage.FindByValue(longUrl); found {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(config.GetFlagHost() + exKey))
		return
	}

	var maxGenerate = 5
	var shortUrl string

	for gen := 0; gen < maxGenerate; gen++ {
		shortUrl = generateID()
		err = h.storage.Save(shortUrl, longUrl)

		if err == nil {
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			_, err = res.Write([]byte(config.GetFlagHost() + shortUrl))
			if err != nil {
				return
			}
			return
		}

		if err.Error() == "Такая ссылка уже есть!" {
			continue
		}

		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusBadRequest)
}

func (h *HandlerStorage) GetUrl(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	inputUrl, err := h.storage.Get(id)
	if err == nil {
		res.Header().Set("Location", inputUrl)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	res.WriteHeader(http.StatusBadRequest)
}
