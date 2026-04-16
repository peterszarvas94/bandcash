package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func main() {
	var (
		flagDBPath  = flag.String("db", "", "SQLite database path (overrides DB_PATH)")
		flagSeedAll = flag.Bool("all", true, "Insert all seed rows")
	)

	flag.Parse()

	dbPath := utils.Env().DBPath
	if *flagDBPath != "" {
		dbPath = *flagDBPath
	}

	absPath, _ := filepath.Abs(dbPath)
	slog.Info("resetting database", "path", absPath)

	pathsToRemove := []string{dbPath, dbPath + "-wal", dbPath + "-shm"}
	for _, path := range pathsToRemove {
		for i := 0; i < 5; i++ {
			err := os.Remove(path)
			if err == nil || os.IsNotExist(err) {
				if err == nil {
					slog.Info("removed file", "path", filepath.Base(path))
				}
				break
			}
			if i < 4 {
				slog.Warn("failed to remove file, retrying", "path", filepath.Base(path), "err", err, "attempt", i+1)
				time.Sleep(100 * time.Millisecond)
			} else {
				slog.Error("failed to remove file after retries", "path", filepath.Base(path), "err", err)
				fmt.Fprintf(os.Stderr, "\nERROR: Cannot delete database file. Make sure the app is not running.\n")
				fmt.Fprintf(os.Stderr, "Run: pkill -f \"air\" && pkill -f \"tmp/server\"\n\n")
				os.Exit(1)
			}
		}
	}

	time.Sleep(100 * time.Millisecond)

	if err := db.Init(dbPath); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database", "err", err)
		}
	}()

	if err := db.Migrate(); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	if !*flagSeedAll {
		slog.Info("no seed action selected")
		return
	}

	if err := db.Seed("heavy_seed.sql"); err != nil {
		slog.Error("failed to run heavy seed", "err", err)
		os.Exit(1)
	}

	const seedGroupID = "grp_SeedDataLabGroup0001"

	var membersCount int
	var eventsCount int
	var expensesCount int
	var participantsCount int
	ctx := context.Background()

	membersCount, err := db.BunDB.NewSelect().TableExpr("members").Where("group_id = ?", seedGroupID).Count(ctx)
	if err != nil {
		slog.Error("failed to count seeded members", "err", err)
		os.Exit(1)
	}
	eventsCount, err = db.BunDB.NewSelect().TableExpr("events").Where("group_id = ?", seedGroupID).Count(ctx)
	if err != nil {
		slog.Error("failed to count seeded events", "err", err)
		os.Exit(1)
	}
	expensesCount, err = db.BunDB.NewSelect().TableExpr("expenses").Where("group_id = ?", seedGroupID).Count(ctx)
	if err != nil {
		slog.Error("failed to count seeded expenses", "err", err)
		os.Exit(1)
	}
	participantsCount, err = db.BunDB.NewSelect().TableExpr("participants").Where("group_id = ?", seedGroupID).Count(ctx)
	if err != nil {
		slog.Error("failed to count seeded participants", "err", err)
		os.Exit(1)
	}

	fmt.Printf("Seeded heavy dataset into %s\n", dbPath)
	fmt.Printf("group_id=%s members=%d events=%d expenses=%d participants=%d\n", seedGroupID, membersCount, eventsCount, expensesCount, participantsCount)
	fmt.Printf("Flag enabled: enable_signup=1\n")
	fmt.Printf("Login examples: peterszarvas94@gmail.com, admin1@bandcash.local, admin2@bandcash.local, admin3@bandcash.local\n")
}
