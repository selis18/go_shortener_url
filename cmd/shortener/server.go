package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/repository"
	"go.uber.org/zap"
)

func InitServer() {
	var database handler.DatabasePinger
	if dsn := config.GetDatabaseDSN(); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			logger.Log.Error("init database error")
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			storage, initErr := repository.NewPostgresStorage(ctx, db)
			cancel()
			if initErr != nil {
				logger.Log.Fatal("init database error", zap.Error(initErr))
			}
			defer db.Close()
			database = db
			handlers := handler.NewHandlerStorage(storage)
			startServer(handlers, database)
			return
		}
	}

	var storage repository.Storage
	filePath := config.GetFileStoragePath()
	if filePath != "" {
		fileStorage, err := repository.NewFileStorage(filePath)
		if err != nil {
			logger.Log.Fatal("init file error", zap.Error(err))
		}
		storage = fileStorage

	} else {
		storage = repository.NewStorageRepo()
	}
	handlers := handler.NewHandlerStorage(storage)
	startServer(handlers, database)
}

func startServer(handlers *handler.HandlerStorage, database handler.DatabasePinger) {
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(gzipMiddleware)
	r.Get("/ping", handler.NewPingHandler(database))
	r.Route("/", func(r chi.Router) {
		r.Use(middleware.AllowContentType("text/plain"))
		r.Get("/{id}", handlers.GetURL)
		r.Post("/", handlers.PostURL)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AllowContentType("application/json"))
		r.Post("/shorten", handlers.PostShorten)
		r.Post("/shorten/batch", handlers.PostShortenBatch)
	})

	err := http.ListenAndServe(config.GetFlagAddress(), r)
	if err != nil {
		panic(err)
	}
}
