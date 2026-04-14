package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"bandcash/internal/db/bunmigrations"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/migrate"

	_ "github.com/mattn/go-sqlite3"
)

var (
	sqlDB *sql.DB
	BunDB *bun.DB
)

//go:embed seeds/*.sql
var seedsFS embed.FS

func Init(dbPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	sqlDB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	BunDB = bun.NewDB(sqlDB, sqlitedialect.New())

	// Set per-connection PRAGMAs
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA cache_size = 10000",
	}
	ctx := context.Background()
	for _, pragma := range pragmas {
		if _, err := BunDB.ExecContext(ctx, pragma); err != nil {
			return fmt.Errorf("failed to set pragma %q: %w", pragma, err)
		}
	}

	slog.Info("database connected", "path", dbPath)
	return nil
}

func Close() error {
	if BunDB != nil {
		if err := BunDB.Close(); err != nil {
			return err
		}
		BunDB = nil
		sqlDB = nil
		return nil
	}
	if sqlDB != nil {
		err := sqlDB.Close()
		sqlDB = nil
		return err
	}
	return nil
}

func Migrate() error {
	ctx := context.Background()
	migrator := migrate.NewMigrator(BunDB, bunmigrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize bun migrations: %w", err)
	}
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to run bun migrations: %w", err)
	}
	if group != nil {
		slog.Info("bun migrations applied", "group_id", group.ID, "migrations", len(group.Migrations))
	}
	return nil
}

func Seed(seedFile string) error {
	if BunDB == nil {
		return fmt.Errorf("database is not initialized")
	}

	content, err := seedsFS.ReadFile("seeds/" + seedFile)
	if err != nil {
		return fmt.Errorf("failed to read seed file %q: %w", seedFile, err)
	}

	if _, err := BunDB.ExecContext(context.Background(), string(content)); err != nil {
		return fmt.Errorf("failed to execute seed file %q: %w", seedFile, err)
	}

	return nil
}
