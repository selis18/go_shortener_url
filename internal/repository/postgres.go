package repository

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/lib/pq"
)

type PostgresStorage struct{ db *sql.DB }

func (s *PostgresStorage) SaveBatch(ctx context.Context, pairs []URLPair) ([]string, error) {
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
		err = tx.QueryRowContext(ctx, `INSERT INTO short_urls(short_url, original_url) VALUES ($1, $2)
			ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
			RETURNING short_url`, pair.ShortURL, pair.OriginalURL).Scan(&keys[i])
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				return nil, ErrShortURLExists
			}
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return keys, nil
}

func NewPostgresStorage(ctx context.Context, db *sql.DB) (*PostgresStorage, error) {
	const schema = `CREATE TABLE IF NOT EXISTS short_urls (short_url TEXT PRIMARY KEY, original_url TEXT NOT NULL UNIQUE)`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, err
	}
	return &PostgresStorage{db: db}, nil
}
func (s *PostgresStorage) Save(k, v string) error { return s.SaveContext(context.Background(), k, v) }
func (s *PostgresStorage) SaveContext(ctx context.Context, k, v string) error {
	if v == "" {
		return ErrShortURLEmpty
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO short_urls(short_url, original_url) VALUES ($1,$2)`, k, v)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrShortURLExists
		}
		return err
	}
	return nil
}
func (s *PostgresStorage) Get(k string) (string, error) { return s.GetContext(context.Background(), k) }
func (s *PostgresStorage) GetContext(ctx context.Context, k string) (string, error) {
	var v string
	err := s.db.QueryRowContext(ctx, `SELECT original_url FROM short_urls WHERE short_url=$1`, k).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrKeyNotFound
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
