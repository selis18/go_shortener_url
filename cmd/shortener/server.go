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
	var storage repository.Storage
	filePath := config.GetFileStoragePath()
	if filePath != "" {
		fileStorage, err := repository.NewFileStorage(filePath)
		if err != nil {
			logger.Log.Fatal("init file error")
		}
		storage = fileStorage

	} else {
		storage = repository.NewStorageRepo()
	}
	handlers := handler.NewHandlerStorage(storage)
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(gzipMiddleware)
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain"))
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.PostURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/shorten", handlers.PostShorten)
	})

	err := http.ListenAndServe(config.GetFlagAddress(), r)
	if err != nil {
		panic(err)
	}
}
