package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/handler"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/repository"
	"github.com/selis18/go_shortener_url/migrations"
	"go.uber.org/zap"
)

func InitServer() {
	var database handler.DatabasePinger
	if dsn := config.GetDatabaseDSN(); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			logger.Log.Fatal("open database error", zap.Error(err))
			return
		}
		defer db.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err != nil {
			logger.Log.Fatal("connect database error", zap.Error(err))
			return
		}
		if err := migrations.Up(db); err != nil {
			logger.Log.Fatal("migrate database error", zap.Error(err))
			return
		}
		storage := repository.NewPostgresStorage(db)
		database = db
		handlers := handler.NewHandlerStorage(storage)
		startServer(handlers, database)
		return
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
	handlers.StartDeletionWorkers()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := handlers.ShutdownDeletes(ctx); err != nil {
			logger.Log.Error("shutdown deletion workers", zap.Error(err))
		}
	}()
	r := chi.NewRouter()

	r.Use(logger.RequestLogger)
	r.Use(gzipMiddleware)
	r.Use(auth.CookieMiddleware)
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
		r.Get("/user/urls", handlers.GetUserURLs)
		r.Delete("/user/urls", handlers.DeleteUserURLs)
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: config.GetFlagAddress(), Handler: r}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("HTTP server", zap.Error(err))
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("shutdown HTTP server", zap.Error(err))
			_ = server.Close()
		}
	}
}
