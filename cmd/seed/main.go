package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"bandcash/internal/db"
	_ "bandcash/internal/utils"
)

type seedEntry struct {
	Title       string
	Time        string
	Description string
	Amount      int64
}

type seedPayee struct {
	Name        string
	Description string
}

type seedParticipant struct {
	EntryIndex int
	PayeeIndex int
	Amount     int64
}

func main() {
	var (
		flagDBPath  = flag.String("db", "", "SQLite database path (overrides DB_PATH)")
		flagSeedAll = flag.Bool("all", true, "Insert all seed rows")
	)

	flag.Parse()

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "sqlite.db"
	}
	if *flagDBPath != "" {
		dbPath = *flagDBPath
	}

	err := os.Remove(dbPath)
	if err != nil && !os.IsNotExist(err) {
		slog.Error("failed to remove existing database", "err", err)
		os.Exit(1)
	}

	err = db.Init(dbPath)
	if err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer func() {
		err := db.Close()
		if err != nil {
			slog.Error("failed to close database", "err", err)
		}
	}()

	err = db.Migrate()
	if err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	if !*flagSeedAll {
		slog.Info("no seed action selected")
		return
	}

	ctx := context.Background()

	// Gigs (entries) - band performances
	entries := []seedEntry{
		{
			Title:       "The Blue Note - Friday Night",
			Time:        time.Now().Add(-240 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Weekend show at downtown jazz club",
			Amount:      80000, // $800.00
		},
		{
			Title:       "Summer Fest Main Stage",
			Time:        time.Now().Add(-192 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Outdoor festival performance",
			Amount:      250000, // $2500.00
		},
		{
			Title:       "Private Wedding - Johnson",
			Time:        time.Now().Add(-144 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Private event, 3-hour set",
			Amount:      120000, // $1200.00
		},
		{
			Title:       "Riverside Bar Gig",
			Time:        time.Now().Add(-96 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Acoustic Wednesday set",
			Amount:      30000, // $300.00
		},
		{
			Title:       "Corporate Event - TechCorp",
			Time:        time.Now().Add(-48 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Holiday party performance",
			Amount:      150000, // $1500.00
		},
		{
			Title:       "Battle of the Bands",
			Time:        time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Competition, won 2nd place",
			Amount:      50000, // $500.00 prize
		},
	}

	// Band members and venues (payees)
	payees := []seedPayee{
		{
			Name:        "Alex Rivera",
			Description: "Lead vocals & rhythm guitar",
		},
		{
			Name:        "Jordan Chen",
			Description: "Lead guitar & backing vocals",
		},
		{
			Name:        "Morgan Blake",
			Description: "Bass guitar",
		},
		{
			Name:        "Casey Martinez",
			Description: "Drums & percussion",
		},
		{
			Name:        "Taylor Kim",
			Description: "Keyboards & synth",
		},
		{
			Name:        "Riley Park",
			Description: "Band manager (10% cut)",
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

	// Participants - band members getting cuts from each gig
	// Index mapping: 0=Alex, 1=Jordan, 2=Morgan, 3=Casey, 4=Taylor, 5=Riley (manager)
	participants := []seedParticipant{
		// The Blue Note - $800 total
		{EntryIndex: 0, PayeeIndex: 0, Amount: 18000}, // Alex: $180
		{EntryIndex: 0, PayeeIndex: 1, Amount: 18000}, // Jordan: $180
		{EntryIndex: 0, PayeeIndex: 2, Amount: 16000}, // Morgan: $160
		{EntryIndex: 0, PayeeIndex: 3, Amount: 16000}, // Casey: $160
		{EntryIndex: 0, PayeeIndex: 4, Amount: 8000},  // Taylor: $80
		{EntryIndex: 0, PayeeIndex: 5, Amount: 4000},  // Riley (manager 10%): $40

		// Summer Fest - $2500 total
		{EntryIndex: 1, PayeeIndex: 0, Amount: 56250}, // Alex: $562.50
		{EntryIndex: 1, PayeeIndex: 1, Amount: 56250}, // Jordan: $562.50
		{EntryIndex: 1, PayeeIndex: 2, Amount: 50000}, // Morgan: $500
		{EntryIndex: 1, PayeeIndex: 3, Amount: 50000}, // Casey: $500
		{EntryIndex: 1, PayeeIndex: 4, Amount: 25000}, // Taylor: $250
		{EntryIndex: 1, PayeeIndex: 5, Amount: 12500}, // Riley (manager 10%): $125

		// Private Wedding - $1200 total
		{EntryIndex: 2, PayeeIndex: 0, Amount: 27000}, // Alex: $270
		{EntryIndex: 2, PayeeIndex: 1, Amount: 27000}, // Jordan: $270
		{EntryIndex: 2, PayeeIndex: 2, Amount: 24000}, // Morgan: $240
		{EntryIndex: 2, PayeeIndex: 3, Amount: 24000}, // Casey: $240
		{EntryIndex: 2, PayeeIndex: 4, Amount: 12000}, // Taylor: $120
		{EntryIndex: 2, PayeeIndex: 5, Amount: 6000},  // Riley (manager 10%): $60

		// Riverside Bar - $300 total (smaller venue, smaller cuts)
		{EntryIndex: 3, PayeeIndex: 0, Amount: 6750}, // Alex: $67.50
		{EntryIndex: 3, PayeeIndex: 1, Amount: 6750}, // Jordan: $67.50
		{EntryIndex: 3, PayeeIndex: 2, Amount: 6000}, // Morgan: $60
		{EntryIndex: 3, PayeeIndex: 3, Amount: 6000}, // Casey: $60
		{EntryIndex: 3, PayeeIndex: 4, Amount: 3000}, // Taylor: $30
		{EntryIndex: 3, PayeeIndex: 5, Amount: 1500}, // Riley (manager 10%): $15

		// Corporate Event - $1500 total
		{EntryIndex: 4, PayeeIndex: 0, Amount: 33750}, // Alex: $337.50
		{EntryIndex: 4, PayeeIndex: 1, Amount: 33750}, // Jordan: $337.50
		{EntryIndex: 4, PayeeIndex: 2, Amount: 30000}, // Morgan: $300
		{EntryIndex: 4, PayeeIndex: 3, Amount: 30000}, // Casey: $300
		{EntryIndex: 4, PayeeIndex: 4, Amount: 15000}, // Taylor: $150
		{EntryIndex: 4, PayeeIndex: 5, Amount: 7500},  // Riley (manager 10%): $75

		// Battle of the Bands - $500 prize
		{EntryIndex: 5, PayeeIndex: 0, Amount: 11250}, // Alex: $112.50
		{EntryIndex: 5, PayeeIndex: 1, Amount: 11250}, // Jordan: $112.50
		{EntryIndex: 5, PayeeIndex: 2, Amount: 10000}, // Morgan: $100
		{EntryIndex: 5, PayeeIndex: 3, Amount: 10000}, // Casey: $100
		{EntryIndex: 5, PayeeIndex: 4, Amount: 5000},  // Taylor: $50
		{EntryIndex: 5, PayeeIndex: 5, Amount: 2500},  // Riley (manager 10%): $25
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
