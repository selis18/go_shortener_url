package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
)

func InitServer() {
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("text/plain"))

	r.Get("/{id}", getUrl)
	r.Post("/", postUrl)

	fmt.Println("Running server on", config.FlagAddress.String())
	err := http.ListenAndServe(config.FlagAddress.String(), r)
	if err != nil {
		panic(err)
	}
}
