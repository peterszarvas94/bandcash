package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

var (
	DB  *sql.DB
	Qry *Queries
)

func Init(dbPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set per-connection PRAGMAs
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA cache_size = 10000",
	}
	for _, pragma := range pragmas {
		if _, err := DB.Exec(pragma); err != nil {
			return fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	// Initialize queries
	Qry = New(DB)

	slog.Info("database connected", "path", dbPath)
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

func Migrate() error {
	err := goose.SetDialect("sqlite3")
	if err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	err = goose.Up(DB, "internal/db/migrations")
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
