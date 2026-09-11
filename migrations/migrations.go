package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var files embed.FS

// Up applies pending embedded migrations without closing the application's pool.
func Up(db *sql.DB) error {
	source, err := iofs.New(files, ".")
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	defer source.Close()
	conn, err := db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Close()
	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("initialize migration driver: %w", err)
	}
	runner, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("initialize migrations: %w", err)
	}
	if err := runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
