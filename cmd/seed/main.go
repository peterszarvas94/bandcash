package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"bandcash/internal/db"
)

type seedEntry struct {
	Title       string
	Time        string
	Description string
	Amount      float64
}

type seedPayee struct {
	Name        string
	Description string
}

type seedParticipant struct {
	EntryIndex int
	PayeeIndex int
	Amount     float64
}

func main() {
	var (
		flagDBPath  = flag.String("db", "", "SQLite database path (overrides DB_PATH)")
		flagSeedAll = flag.Bool("all", true, "Insert all seed rows")
	)
	flag.Parse()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./sqlite.db"
	}
	if *flagDBPath != "" {
		dbPath = *flagDBPath
	}

	if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
		slog.Error("failed to remove existing database", "err", err)
		os.Exit(1)
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

	ctx := context.Background()
	entries := []seedEntry{
		{
			Title:       "Client invoice",
			Time:        time.Now().Add(-72 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Invoice for February retainers",
			Amount:      1250.00,
		},
		{
			Title:       "Studio rent",
			Time:        time.Now().Add(-48 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Monthly studio rent",
			Amount:      -950.00,
		},
		{
			Title:       "Equipment",
			Time:        time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Mic stand and cables",
			Amount:      -85.50,
		},
	}

	payees := []seedPayee{
		{
			Name:        "Northside Studios",
			Description: "Shared rehearsal space",
		},
		{
			Name:        "Ari Lane",
			Description: "Session guitarist",
		},
		{
			Name:        "Soundcheck Supply",
			Description: "Gear and cables",
		},
	}

	createdEntries := make([]db.Entry, 0, len(entries))
	for _, entry := range entries {
		created, err := db.Qry.CreateEntry(ctx, db.CreateEntryParams{
			Title:       entry.Title,
			Time:        entry.Time,
			Description: entry.Description,
			Amount:      entry.Amount,
		})
		if err != nil {
			slog.Error("failed to insert seed entry", "title", entry.Title, "err", err)
			os.Exit(1)
		}
		createdEntries = append(createdEntries, created)
	}

	createdPayees := make([]db.Payee, 0, len(payees))
	for _, payee := range payees {
		created, err := db.Qry.CreatePayee(ctx, db.CreatePayeeParams{
			Name:        payee.Name,
			Description: payee.Description,
		})
		if err != nil {
			slog.Error("failed to insert seed payee", "name", payee.Name, "err", err)
			os.Exit(1)
		}
		createdPayees = append(createdPayees, created)
	}

	participants := []seedParticipant{
		{EntryIndex: 0, PayeeIndex: 1, Amount: 400.00},
		{EntryIndex: 1, PayeeIndex: 0, Amount: -450.00},
		{EntryIndex: 2, PayeeIndex: 2, Amount: -85.50},
	}

	for _, participant := range participants {
		entry := createdEntries[participant.EntryIndex]
		payee := createdPayees[participant.PayeeIndex]
		_, err := db.Qry.AddParticipant(ctx, db.AddParticipantParams{
			EntryID: entry.ID,
			PayeeID: payee.ID,
			Amount:  participant.Amount,
		})
		if err != nil {
			slog.Error("failed to insert seed participant", "entry_id", entry.ID, "payee_id", payee.ID, "err", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Seeded %d entries, %d payees, %d participants into %s\n", len(entries), len(payees), len(participants), dbPath)
}
