package handler

import (
	"context"
	"net/http"
	"time"
)

// DatabasePinger is implemented by *sql.DB.
type DatabasePinger interface {
	PingContext(context.Context) error
}

func NewPingHandler(db DatabasePinger) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if db == nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
