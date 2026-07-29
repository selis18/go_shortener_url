package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func InitServer() {
	storage := repository.NewStorageRepo()
	handlers := handler.NewHandlerStorage(storage)
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("text/plain"))

	r.Get("/{id}", handlers.GetUrl)
	r.Post("/", handlers.PostUrl)

	err := http.ListenAndServe(config.GetFlagAdress(), r)
	if err != nil {
		panic(err)
	}
}
