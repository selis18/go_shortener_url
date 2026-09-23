package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/selis18/go_shortener_url/internal/auth"
	"github.com/selis18/go_shortener_url/internal/logger"
	"github.com/selis18/go_shortener_url/internal/repository"
	"go.uber.org/zap"
)

func (h *HandlerStorage) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	defer r.Body.Close()
	var ids []string
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&ids); err != nil || ids == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, id := range ids {
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	storage, ok := h.storage.(repository.DeleteStorage)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.deleteMu.Lock()
	for _, id := range ids {
		h.pendingDeletes = append(h.pendingDeletes, repository.DeleteRequest{UserID: userID, ShortURL: id})
	}
	if !h.deleting && len(h.pendingDeletes) > 0 {
		h.deleting = true
		go h.runDeletes(storage)
	}
	h.deleteMu.Unlock()
	w.WriteHeader(http.StatusAccepted)
}

func (h *HandlerStorage) runDeletes(storage repository.DeleteStorage) {
	for {
		time.Sleep(25 * time.Millisecond)
		h.deleteMu.Lock()
		if len(h.pendingDeletes) == 0 {
			h.pendingDeletes = nil
			h.deleting = false
			h.deleteMu.Unlock()
			return
		}
		n := min(len(h.pendingDeletes), 1000)
		batch := h.pendingDeletes[:n:n]
		h.pendingDeletes = h.pendingDeletes[n:]
		h.deleteMu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := storage.DeleteBatch(ctx, batch)
		cancel()
		if err != nil {
			logger.Log.Error("delete URLs", zap.Error(err))
		}
	}
}
