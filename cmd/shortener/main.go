package main

import (
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
	"strings"
)

type userURL struct {
	inputUrl string
	shortUrl string
}

var storage = map[string]string{}

const form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/text/plain" method="post">
            <label>Ссылка <input type="text" name="longUrl"></label>
            <input type="submit" value="Sent">
        </form>
    </body>
</html>`

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

func mainPage(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		res.WriteHeader(http.StatusCreated)
		url := userURL{
			req.FormValue("longUrl"),
			generateID(),
		}
		storage[url.inputUrl] = url.shortUrl
		io.WriteString(res, "http://localhost:8080/"+storage[url.inputUrl])
	} else {
		io.WriteString(res, form)
		res.WriteHeader(http.StatusBadRequest)
	}
}

func getUrl(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		urlPaths := strings.Split(req.URL.Path, "/")
		id := urlPaths[3]
		for k, v := range storage {
			if v == id {
				http.Redirect(res, req, k, http.StatusTemporaryRedirect)
			}
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/text/plain", mainPage)
	mux.HandleFunc("/text/plain/", getUrl)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
