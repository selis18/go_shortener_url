package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/lib/pq"
	"github.com/selis18/go_shortener_url/internal/model"
)

type DeleteRequest struct {
	UserID   string
	ShortURL string
}

type DeleteStorage interface {
	DeleteBatch(context.Context, []DeleteRequest) error
}

func (s *StorageRepo) DeleteBatch(ctx context.Context, requests []DeleteRequest) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, r := range requests {
		if r.UserID != "" && s.owners[r.ShortURL] == r.UserID {
			s.deleted[r.ShortURL] = true
		}
	}
	return nil
}

func (s *PostgresStorage) DeleteBatch(ctx context.Context, requests []DeleteRequest) error {
	if len(requests) == 0 {
		return ctx.Err()
	}
	users, keys := make([]string, len(requests)), make([]string, len(requests))
	for i, r := range requests {
		users[i], keys[i] = r.UserID, r.ShortURL
	}
	_, err := s.db.ExecContext(ctx, `UPDATE short_urls s SET is_deleted = TRUE
 FROM unnest($1::text[], $2::text[]) AS d(user_id, short_url)
 WHERE s.short_url = d.short_url AND s.user_id = d.user_id
 AND d.user_id <> '' AND NOT s.is_deleted`, pq.Array(users), pq.Array(keys))
	return err
}

func (s *FileStorage) DeleteBatch(ctx context.Context, requests []DeleteRequest) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.storage.mutex.Lock()
	defer s.storage.mutex.Unlock()
	var data bytes.Buffer
	encoder := json.NewEncoder(&data)
	keys := make(map[string]bool)
	for _, r := range requests {
		if err := ctx.Err(); err != nil {
			return err
		}
		if r.UserID == "" || s.storage.owners[r.ShortURL] != r.UserID || s.storage.deleted[r.ShortURL] || keys[r.ShortURL] {
			continue
		}
		if err := encoder.Encode(model.JSONStorage{UUID: strconv.Itoa(s.nextUUID + len(keys)), UserID: r.UserID,
			ShortURL: r.ShortURL, OriginalURL: s.storage.storage[r.ShortURL], DeletedFlag: true}); err != nil {
			return err
		}
		keys[r.ShortURL] = true
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.appendBatch(data.Bytes()); err != nil {
		return err
	}
	for key := range keys {
		s.storage.deleted[key] = true
	}
	s.nextUUID += len(keys)
	return nil
}
