package main

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

type userUrl struct {
	inputUrl string
}

func newUserUrl(i string) userUrl {
	return userUrl{i}
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

func postUrl(res http.ResponseWriter, req *http.Request) {
	url := userUrl{
		inputUrl: strings.TrimSpace(req.FormValue("longUrl")),
	}

	if url.inputUrl == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	shortUrl := generateID()
	err := StorageR.Save(shortUrl, url)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	render.PlainText(res, req, "http://localhost:8080/"+shortUrl)
}

func getUrl(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	for k, v := range StorageR.storage {
		if k == id {
			//http.Redirect(res, req, v.inputUrl, http.StatusTemporaryRedirect)
			res.Header().Set("Location", v.inputUrl)
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
	res.WriteHeader(http.StatusBadRequest)
}
