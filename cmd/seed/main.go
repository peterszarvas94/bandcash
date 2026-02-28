package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func main() {
	var (
		flagDBPath  = flag.String("db", "", "SQLite database path (overrides DB_PATH)")
		flagSeedAll = flag.Bool("all", true, "Insert all seed rows")
	)

	flag.Parse()

	utils.LoadAppDotEnv()

	dbPath := utils.Env().DBPath
	if *flagDBPath != "" {
		dbPath = *flagDBPath
	}

	pathsToRemove := []string{dbPath, dbPath + "-wal", dbPath + "-shm"}
	for _, path := range pathsToRemove {
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			slog.Error("failed to remove existing database file", "path", path, "err", err)
			os.Exit(1)
		}
	}

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

	if err := db.DB.QueryRow("SELECT COUNT(*) FROM members WHERE group_id = ?", seedGroupID).Scan(&membersCount); err != nil {
		slog.Error("failed to count seeded members", "err", err)
		os.Exit(1)
	}
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM events WHERE group_id = ?", seedGroupID).Scan(&eventsCount); err != nil {
		slog.Error("failed to count seeded events", "err", err)
		os.Exit(1)
	}
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM expenses WHERE group_id = ?", seedGroupID).Scan(&expensesCount); err != nil {
		slog.Error("failed to count seeded expenses", "err", err)
		os.Exit(1)
	}
	if err := db.DB.QueryRow("SELECT COUNT(*) FROM participants WHERE group_id = ?", seedGroupID).Scan(&participantsCount); err != nil {
		slog.Error("failed to count seeded participants", "err", err)
		os.Exit(1)
	}

	fmt.Printf("Seeded heavy dataset into %s\n", dbPath)
	fmt.Printf("group_id=%s members=%d events=%d expenses=%d participants=%d\n", seedGroupID, membersCount, eventsCount, expensesCount, participantsCount)
	fmt.Printf("Flag enabled: enable_signup=1\n")
	fmt.Printf("Login examples: superadmin@bandcash.local, admin1@bandcash.local, admin2@bandcash.local, admin3@bandcash.local\n")
}
