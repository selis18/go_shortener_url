package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func InitServer() {
	storage := repository.NewStorageRepo()
	handlers := handler.NewHandlerStorage(storage)
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("text/plain"))
	r.Use(logger.RequestLogger)

	r.Get("/{id}", handlers.GetURL)
	r.Post("/", handlers.PostURL)

	err := http.ListenAndServe(config.GetFlagAdress(), r)
	if err != nil {
		panic(err)
	}
}
