package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitServer() {
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("text/plain"))

	r.Get("/{id}", getUrl)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	r.Post("/", postUrl)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
