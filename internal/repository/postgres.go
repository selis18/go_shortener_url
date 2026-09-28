package repository

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/lib/pq"
)

type PostgresStorage struct{ db *sql.DB }

func (s *PostgresStorage) GetUserURLs(ctx context.Context, userID string) ([]URLPair, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT s.short_url, s.original_url
		FROM short_urls s JOIN user_urls u ON u.short_url = s.short_url
		WHERE u.user_id = $1 AND s.is_deleted = FALSE ORDER BY s.short_url`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	urls := make([]URLPair, 0)
	for rows.Next() {
		var pair URLPair
		if err := rows.Scan(&pair.ShortURL, &pair.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, pair)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *PostgresStorage) SaveBatch(ctx context.Context, pairs []URLPair, userID string) ([]string, error) {
	keys := make([]string, len(pairs))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return keys, nil
	}
	for _, pair := range pairs {
		if pair.OriginalURL == "" {
			return nil, ErrShortURLEmpty
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	order := make([]int, len(pairs))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool { return pairs[order[i]].OriginalURL < pairs[order[j]].OriginalURL })
	for _, i := range order {
		pair := pairs[i]
		err = tx.QueryRowContext(ctx, `INSERT INTO short_urls(short_url, original_url, user_id) VALUES ($1, $2, $3)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING short_url`, pair.ShortURL, pair.OriginalURL, userID).Scan(&keys[i])
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return nil, ErrShortURLExists
			}
			return nil, err
		}
		if userID != "" {
			if _, err := tx.ExecContext(ctx, "INSERT INTO user_urls(user_id, short_url) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, keys[i]); err != nil {
				return nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return keys, nil
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}
func (s *PostgresStorage) Save(k, v, userID string) error {
	return s.SaveContext(context.Background(), k, v, userID)
}
func (s *PostgresStorage) SaveContext(ctx context.Context, k, v, userID string) error {
	if v == "" {
		return ErrShortURLEmpty
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var key string
	err = tx.QueryRowContext(ctx, "INSERT INTO short_urls(short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (original_url) DO NOTHING RETURNING short_url", k, v, userID).Scan(&key)
	conflict := errors.Is(err, sql.ErrNoRows)
	if conflict {
		err = tx.QueryRowContext(ctx, "SELECT short_url FROM short_urls WHERE original_url=$1", v).Scan(&key)
	}
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrShortURLExists
		}
		return err
	}
	if userID != "" {
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_urls(user_id, short_url) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, key); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if conflict {
		return ErrConflict
	}
	return nil
}

func (s *PostgresStorage) Get(k string) (string, error) { return s.GetContext(context.Background(), k) }
func (s *PostgresStorage) GetContext(ctx context.Context, k string) (string, error) {
	var v string
	var deleted bool
	err := s.db.QueryRowContext(ctx, `SELECT original_url, is_deleted FROM short_urls WHERE short_url=$1`, k).Scan(&v, &deleted)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrKeyNotFound
	}
	if err == nil && deleted {
		return "", ErrDeleted
	}
	return v, err
}
func (s *PostgresStorage) FindByValue(v string) (string, bool) {
	k, ok, _ := s.FindByValueContext(context.Background(), v)
	return k, ok
}
func (s *PostgresStorage) FindByValueContext(ctx context.Context, v string) (string, bool, error) {
	var k string
	err := s.db.QueryRowContext(ctx, `SELECT short_url FROM short_urls WHERE original_url=$1`, v).Scan(&k)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return k, err == nil, err
}
