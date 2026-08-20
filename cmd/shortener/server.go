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

	r.Use(logger.RequestLogger)
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain"))
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.PostURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/shorten", handlers.PostShorten)
	})

	err := http.ListenAndServe(config.GetFlagAdress(), r)
	if err != nil {
		panic(err)
	}
}
