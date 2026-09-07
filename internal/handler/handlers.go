package handler

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
	"go.uber.org/zap"
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

func (h *HandlerStorage) generateShortURL(URL string) (string, error) {
	const maxGenerate = 5
	for gen := 0; gen < maxGenerate; gen++ {
		shortURL := generateID()
		err := h.storage.Save(shortURL, URL)
		//использовала ИИ для помощи с условием ошибки
		if errors.Is(err, repository.ErrShortURLExists) {
			continue
		}
		if err != nil {
			return "", err
		}
		return config.GetFlagHost() + shortURL, nil
	}
	return "", fmt.Errorf("links were not generated after %d attempts", maxGenerate)
}

func (h *HandlerStorage) PostURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	longURL := strings.TrimSpace(string(body))
	if longURL == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if exKey, found := h.storage.FindByValue(longURL); found {
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(config.GetFlagHost() + exKey))
		return
	}

	var shortURL string
	shortURL, err = h.generateShortURL(longURL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	if _, err := res.Write([]byte(shortURL)); err != nil {
		return
	}
}

func (h *HandlerStorage) GetURL(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	inputURL, err := h.storage.Get(id)
	if err == nil {
		res.Header().Set("Location", inputURL)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	res.WriteHeader(http.StatusBadRequest)
}

func (h *HandlerStorage) PostShorten(res http.ResponseWriter, req *http.Request) {
	var request model.Request
	var err error

	decoder := json.NewDecoder(req.Body)
	if err = decoder.Decode(&request); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	var shortURL string

	if shortURL, err = h.generateShortURL(request.URL); err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := model.Response{
		ShortURL: shortURL,
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	encoder := json.NewEncoder(res)
	if err = encoder.Encode(&response); err != nil {
		logger.Log.Debug("error encoding response", zap.Error(err))
		return
	}
}
