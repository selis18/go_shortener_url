package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/selis18/go_shortener_url/internal/config"
	"github.com/selis18/go_shortener_url/internal/model"
	"github.com/selis18/go_shortener_url/internal/repository"
)

func (h *HandlerStorage) PostShortenBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var request []model.BatchRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil || len(request) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pairs := make([]repository.URLPair, len(request))
	for i, item := range request {
		if strings.TrimSpace(item.OriginalURL) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pairs[i].OriginalURL = item.OriginalURL
	}
	storage, ok := h.storage.(repository.BatchStorage)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var keys []string
	var err error
	for attempt := 0; attempt < 5; attempt++ {
		for i := range pairs {
			pairs[i].ShortURL = generateID()
		}
		keys, err = storage.SaveBatch(r.Context(), pairs)
		if !errors.Is(err, repository.ErrShortURLExists) {
			break
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := make([]model.BatchResponse, len(request))
	for i, item := range request {
		response[i] = model.BatchResponse{CorrelationID: item.CorrelationID, ShortURL: config.GetFlagHost() + keys[i]}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
