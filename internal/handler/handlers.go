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

var longUrl string

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

func PostUrl(res http.ResponseWriter, req *http.Request) {
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
	shortUrl := generateID()
	err = repository.StorageR.Save(shortUrl, longUrl)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(config.FlagHost + shortUrl))
	if err != nil {
		return
	}
}

func GetUrl(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if _, e := repository.StorageR.Storage[id]; e {
		res.Header().Set("Location", repository.StorageR.Storage[id])
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	res.WriteHeader(http.StatusBadRequest)
}
