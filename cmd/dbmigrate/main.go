package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"bandcash/internal/db"
	"bandcash/internal/db/bunmigrations"
	"bandcash/internal/utils"

	"github.com/uptrace/bun/migrate"
)

func main() {
	var (
		cmd  = flag.String("cmd", "up", "Migration command: up, down, status, create")
		name = flag.String("name", "", "Migration name when cmd=create")
	)
	flag.Parse()

	utils.SetupLogger()

	if err := db.Init(utils.Env().DBPath); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "err", err)
		}
	}()

	ctx := context.Background()
	migrator := migrate.NewMigrator(db.BunDB, bunmigrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		slog.Error("failed to initialize bun migrations", "err", err)
		os.Exit(1)
	}

	switch *cmd {
	case "up":
		group, err := migrator.Migrate(ctx)
		if err != nil {
			slog.Error("migration up failed", "err", err)
			os.Exit(1)
		}
		if group == nil {
			fmt.Println("No new migrations")
			return
		}
		fmt.Printf("Applied migration group %d (%d migrations)\n", group.ID, len(group.Migrations))
	case "down":
		group, err := migrator.Rollback(ctx)
		if err != nil {
			slog.Error("migration down failed", "err", err)
			os.Exit(1)
		}
		if group == nil || len(group.Migrations) == 0 {
			fmt.Println("No migration group to rollback")
			return
		}
		fmt.Printf("Rolled back migration group %d (%d migrations)\n", group.ID, len(group.Migrations))
	case "status":
		migrationsWithStatus, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			slog.Error("failed to get migration status", "err", err)
			os.Exit(1)
		}
		for _, mig := range migrationsWithStatus {
			status := "pending"
			if mig.IsApplied() {
				status = "applied"
			}
			fmt.Printf("%s\t%s\n", status, mig.Name)
		}
	case "create":
		if *name == "" {
			fmt.Println("-name is required when -cmd=create")
			os.Exit(1)
		}
		files, err := migrator.CreateSQLMigrations(ctx, *name)
		if err != nil {
			slog.Error("failed to create migration files", "err", err)
			os.Exit(1)
		}
		for _, f := range files {
			fmt.Println(f.Path)
		}
	default:
		fmt.Println("unknown command, use: up, down, status, create")
		os.Exit(1)
	}
}
