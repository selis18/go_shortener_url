package main

import (
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
)

type userUrl struct {
	inputUrl string
	shortUrl string
}

func newUserUrl(i string, s string) userUrl {
	return userUrl{i, s}
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
	if req.Method == http.MethodPost && req.FormValue("longUrl") != "" && req.FormValue("longUrl") != " " {
		url := newUserUrl(
			req.FormValue("longUrl"),
			generateID(),
		)
		err := StorageR.Save(url.shortUrl, url)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, "http://localhost:8080/"+url.shortUrl)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func getUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		id := req.URL.Path[1:]
		if id == "" {
			res.WriteHeader(http.StatusBadRequest)
		}
		for k, v := range StorageR.storage {
			if k == id {
				//http.Redirect(res, req, v.inputUrl, http.StatusTemporaryRedirect)
				res.Header().Set("Location", v.inputUrl)
				res.WriteHeader(http.StatusTemporaryRedirect)
			}
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
