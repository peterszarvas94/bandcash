package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type seedEvent struct {
	Title       string
	Time        string
	Description string
	Amount      int64
}

type seedMember struct {
	Name        string
	Description string
}

type seedParticipant struct {
	EventIndex  int
	MemberIndex int
	Amount      int64
}

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

	err := db.Init(dbPath)
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

	// 1. Create admin user
	adminUser, err := db.Qry.CreateUser(ctx, db.CreateUserParams{
		ID:    utils.GenerateID("usr"),
		Email: "admin@bandcash.local",
	})
	if err != nil {
		slog.Error("failed to create admin user", "err", err)
		os.Exit(1)
	}
	slog.Info("created admin user", "id", adminUser.ID, "email", adminUser.Email)

	// 2. Create group
	group, err := db.Qry.CreateGroup(ctx, db.CreateGroupParams{
		ID:          utils.GenerateID("grp"),
		Name:        "My Band",
		AdminUserID: adminUser.ID,
	})
	if err != nil {
		slog.Error("failed to create group", "err", err)
		os.Exit(1)
	}
	slog.Info("created group", "id", group.ID, "name", group.Name)

	groupID := group.ID

	// Koncertek (events) - zenekari fellépések
	events := []seedEvent{
		{
			Title:       "Blue Note - Péntek este",
			Time:        time.Now().Add(-240 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Hétvégi fellépés a belvárosi jazz klubban",
			Amount:      80000, // $800.00
		},
		{
			Title:       "Nyári Fesztivál - Nagyszínpad",
			Time:        time.Now().Add(-192 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Szabadtéri fesztivál fellépés",
			Amount:      250000, // $2500.00
		},
		{
			Title:       "Privát esküvő - Johnson",
			Time:        time.Now().Add(-144 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Privát rendezvény, 3 órás műsor",
			Amount:      120000, // $1200.00
		},
		{
			Title:       "Riverside Bar koncert",
			Time:        time.Now().Add(-96 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Akusztikus szerda esti műsor",
			Amount:      30000, // $300.00
		},
		{
			Title:       "Céges rendezvény - TechCorp",
			Time:        time.Now().Add(-48 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Év végi céges buli",
			Amount:      150000, // $1500.00
		},
		{
			Title:       "Zenekarok csatája",
			Time:        time.Now().Add(-24 * time.Hour).Format("2006-01-02T15:04"),
			Description: "Verseny, 2. helyezés",
			Amount:      50000, // $500.00 prize
		},
	}

	// Zenekari tagok és helyszínek (members)
	members := []seedMember{
		{
			Name:        "Nagy Áron",
			Description: "Ének, ritmusgitár",
		},
		{
			Name:        "Kiss Zsófi",
			Description: "Szólógitár, vokál",
		},
		{
			Name:        "Tóth Márk",
			Description: "Basszusgitár",
		},
		{
			Name:        "Varga Lilla",
			Description: "Dob, ütőhangszerek",
		},
		{
			Name:        "Fekete Bence",
			Description: "Billentyűk, szinti",
		},
		{
			Name:        "Szabó Réka",
			Description: "Menedzser (10% részesedés)",
		},
	}

	createdEvents := make([]db.Event, 0, len(events))
	for _, event := range events {
		created, err := db.Qry.CreateEvent(ctx, db.CreateEventParams{
			ID:          utils.GenerateID(utils.PrefixEvent),
			GroupID:     groupID,
			Title:       event.Title,
			Time:        event.Time,
			Description: event.Description,
			Amount:      event.Amount,
		})
		if err != nil {
			slog.Error("failed to insert seed event", "title", event.Title, "err", err)
			os.Exit(1)
		}
		createdEvents = append(createdEvents, created)
	}

	createdMembers := make([]db.Member, 0, len(members))
	for _, member := range members {
		created, err := db.Qry.CreateMember(ctx, db.CreateMemberParams{
			ID:          utils.GenerateID(utils.PrefixMember),
			GroupID:     groupID,
			Name:        member.Name,
			Description: member.Description,
		})
		if err != nil {
			slog.Error("failed to insert seed member", "name", member.Name, "err", err)
			os.Exit(1)
		}
		createdMembers = append(createdMembers, created)
	}

	// Résztvevők - zenekari tagok részesedése koncertenként
	// Index mapping: 0=Áron, 1=Zsófi, 2=Márk, 3=Lilla, 4=Bence, 5=Réka (menedzser)
	participants := []seedParticipant{
		// Blue Note - $800 total
		{EventIndex: 0, MemberIndex: 0, Amount: 18000}, // Áron: $180
		{EventIndex: 0, MemberIndex: 1, Amount: 18000}, // Zsófi: $180
		{EventIndex: 0, MemberIndex: 2, Amount: 16000}, // Márk: $160
		{EventIndex: 0, MemberIndex: 3, Amount: 16000}, // Lilla: $160
		{EventIndex: 0, MemberIndex: 4, Amount: 8000},  // Bence: $80
		{EventIndex: 0, MemberIndex: 5, Amount: 4000},  // Réka (menedzser 10%): $40

		// Nyári Fesztivál - $2500 total
		{EventIndex: 1, MemberIndex: 0, Amount: 56250}, // Áron: $562.50
		{EventIndex: 1, MemberIndex: 1, Amount: 56250}, // Zsófi: $562.50
		{EventIndex: 1, MemberIndex: 2, Amount: 50000}, // Márk: $500
		{EventIndex: 1, MemberIndex: 3, Amount: 50000}, // Lilla: $500
		{EventIndex: 1, MemberIndex: 4, Amount: 25000}, // Bence: $250
		{EventIndex: 1, MemberIndex: 5, Amount: 12500}, // Réka (menedzser 10%): $125

		// Privát esküvő - $1200 total
		{EventIndex: 2, MemberIndex: 0, Amount: 27000}, // Áron: $270
		{EventIndex: 2, MemberIndex: 1, Amount: 27000}, // Zsófi: $270
		{EventIndex: 2, MemberIndex: 2, Amount: 24000}, // Márk: $240
		{EventIndex: 2, MemberIndex: 3, Amount: 24000}, // Lilla: $240
		{EventIndex: 2, MemberIndex: 4, Amount: 12000}, // Bence: $120
		{EventIndex: 2, MemberIndex: 5, Amount: 6000},  // Réka (menedzser 10%): $60

		// Riverside Bar - $300 total (kisebb hely, kisebb részesedés)
		{EventIndex: 3, MemberIndex: 0, Amount: 6750}, // Áron: $67.50
		{EventIndex: 3, MemberIndex: 1, Amount: 6750}, // Zsófi: $67.50
		{EventIndex: 3, MemberIndex: 2, Amount: 6000}, // Márk: $60
		{EventIndex: 3, MemberIndex: 3, Amount: 6000}, // Lilla: $60
		{EventIndex: 3, MemberIndex: 4, Amount: 3000}, // Bence: $30
		{EventIndex: 3, MemberIndex: 5, Amount: 1500}, // Réka (menedzser 10%): $15

		// Céges rendezvény - $1500 total
		{EventIndex: 4, MemberIndex: 0, Amount: 33750}, // Áron: $337.50
		{EventIndex: 4, MemberIndex: 1, Amount: 33750}, // Zsófi: $337.50
		{EventIndex: 4, MemberIndex: 2, Amount: 30000}, // Márk: $300
		{EventIndex: 4, MemberIndex: 3, Amount: 30000}, // Lilla: $300
		{EventIndex: 4, MemberIndex: 4, Amount: 15000}, // Bence: $150
		{EventIndex: 4, MemberIndex: 5, Amount: 7500},  // Réka (menedzser 10%): $75

		// Zenekarok csatája - $500 prize
		{EventIndex: 5, MemberIndex: 0, Amount: 11250}, // Áron: $112.50
		{EventIndex: 5, MemberIndex: 1, Amount: 11250}, // Zsófi: $112.50
		{EventIndex: 5, MemberIndex: 2, Amount: 10000}, // Márk: $100
		{EventIndex: 5, MemberIndex: 3, Amount: 10000}, // Lilla: $100
		{EventIndex: 5, MemberIndex: 4, Amount: 5000},  // Bence: $50
		{EventIndex: 5, MemberIndex: 5, Amount: 2500},  // Réka (menedzser 10%): $25
	}

	for _, participant := range participants {
		event := createdEvents[participant.EventIndex]
		member := createdMembers[participant.MemberIndex]
		_, err := db.Qry.AddParticipant(ctx, db.AddParticipantParams{
			GroupID:  groupID,
			EventID:  event.ID,
			MemberID: member.ID,
			Amount:   participant.Amount,
		})
		if err != nil {
			slog.Error("failed to insert seed participant", "event_id", event.ID, "member_id", member.ID, "err", err)
			os.Exit(1)
		}
	}

	viewerFixtures := []struct {
		OwnerEmail   string
		GroupName    string
		Events       []seedEvent
		Members      []seedMember
		Participants []seedParticipant
	}{
		{
			OwnerEmail: "owner.one@bandcash.local",
			GroupName:  "Road Crew Collective",
			Events: []seedEvent{
				{Title: "Warehouse Rehearsal", Time: time.Now().Add(-72 * time.Hour).Format("2006-01-02T15:04"), Description: "Paid technical rehearsal", Amount: 60000},
				{Title: "Downtown Showcase", Time: time.Now().Add(-12 * time.Hour).Format("2006-01-02T15:04"), Description: "Support slot in city center", Amount: 90000},
			},
			Members: []seedMember{
				{Name: "Crew Alice", Description: "Lead vocals"},
				{Name: "Crew Bob", Description: "Drums"},
				{Name: "Crew Cara", Description: "Sound tech"},
			},
			Participants: []seedParticipant{
				{EventIndex: 0, MemberIndex: 0, Amount: 25000},
				{EventIndex: 0, MemberIndex: 1, Amount: 20000},
				{EventIndex: 0, MemberIndex: 2, Amount: 10000},
				{EventIndex: 1, MemberIndex: 0, Amount: 35000},
				{EventIndex: 1, MemberIndex: 1, Amount: 30000},
				{EventIndex: 1, MemberIndex: 2, Amount: 15000},
			},
		},
		{
			OwnerEmail: "owner.two@bandcash.local",
			GroupName:  "Late Night Session",
			Events: []seedEvent{
				{Title: "Jazz Basement", Time: time.Now().Add(-96 * time.Hour).Format("2006-01-02T15:04"), Description: "Ticketed evening set", Amount: 70000},
				{Title: "Studio Overdub", Time: time.Now().Add(-6 * time.Hour).Format("2006-01-02T15:04"), Description: "Paid recording session", Amount: 50000},
			},
			Members: []seedMember{
				{Name: "Session Dani", Description: "Keyboard"},
				{Name: "Session Erik", Description: "Bass"},
				{Name: "Session Faye", Description: "Backing vocals"},
			},
			Participants: []seedParticipant{
				{EventIndex: 0, MemberIndex: 0, Amount: 25000},
				{EventIndex: 0, MemberIndex: 1, Amount: 22000},
				{EventIndex: 0, MemberIndex: 2, Amount: 13000},
				{EventIndex: 1, MemberIndex: 0, Amount: 18000},
				{EventIndex: 1, MemberIndex: 1, Amount: 17000},
				{EventIndex: 1, MemberIndex: 2, Amount: 10000},
			},
		},
	}

	totalUsers := 1
	totalGroups := 1
	totalEvents := len(events)
	totalMembers := len(members)
	totalParticipants := len(participants)
	totalViewerLinks := 0

	for _, fixture := range viewerFixtures {
		ownerUser, err := db.Qry.CreateUser(ctx, db.CreateUserParams{
			ID:    utils.GenerateID("usr"),
			Email: fixture.OwnerEmail,
		})
		if err != nil {
			slog.Error("failed to create fixture owner user", "email", fixture.OwnerEmail, "err", err)
			os.Exit(1)
		}

		viewerGroup, err := db.Qry.CreateGroup(ctx, db.CreateGroupParams{
			ID:          utils.GenerateID("grp"),
			Name:        fixture.GroupName,
			AdminUserID: ownerUser.ID,
		})
		if err != nil {
			slog.Error("failed to create viewer fixture group", "group_name", fixture.GroupName, "err", err)
			os.Exit(1)
		}

		_, err = db.Qry.CreateGroupReader(ctx, db.CreateGroupReaderParams{
			ID:      utils.GenerateID("grd"),
			UserID:  adminUser.ID,
			GroupID: viewerGroup.ID,
		})
		if err != nil {
			slog.Error("failed to add admin user as group viewer", "group_id", viewerGroup.ID, "err", err)
			os.Exit(1)
		}

		viewerEvents := make([]db.Event, 0, len(fixture.Events))
		for _, event := range fixture.Events {
			createdEvent, err := db.Qry.CreateEvent(ctx, db.CreateEventParams{
				ID:          utils.GenerateID(utils.PrefixEvent),
				GroupID:     viewerGroup.ID,
				Title:       event.Title,
				Time:        event.Time,
				Description: event.Description,
				Amount:      event.Amount,
			})
			if err != nil {
				slog.Error("failed to create viewer fixture event", "group_id", viewerGroup.ID, "title", event.Title, "err", err)
				os.Exit(1)
			}
			viewerEvents = append(viewerEvents, createdEvent)
		}

		viewerMembers := make([]db.Member, 0, len(fixture.Members))
		for _, member := range fixture.Members {
			createdMember, err := db.Qry.CreateMember(ctx, db.CreateMemberParams{
				ID:          utils.GenerateID(utils.PrefixMember),
				GroupID:     viewerGroup.ID,
				Name:        member.Name,
				Description: member.Description,
			})
			if err != nil {
				slog.Error("failed to create viewer fixture member", "group_id", viewerGroup.ID, "name", member.Name, "err", err)
				os.Exit(1)
			}
			viewerMembers = append(viewerMembers, createdMember)
		}

		for _, participant := range fixture.Participants {
			event := viewerEvents[participant.EventIndex]
			member := viewerMembers[participant.MemberIndex]
			_, err := db.Qry.AddParticipant(ctx, db.AddParticipantParams{
				GroupID:  viewerGroup.ID,
				EventID:  event.ID,
				MemberID: member.ID,
				Amount:   participant.Amount,
			})
			if err != nil {
				slog.Error("failed to create viewer fixture participant", "group_id", viewerGroup.ID, "event_id", event.ID, "member_id", member.ID, "err", err)
				os.Exit(1)
			}
		}

		totalUsers++
		totalGroups++
		totalEvents += len(fixture.Events)
		totalMembers += len(fixture.Members)
		totalParticipants += len(fixture.Participants)
		totalViewerLinks++
	}

	fmt.Printf("Seeded %d users, %d groups, %d events, %d members, %d participants, %d viewer links into %s\n", totalUsers, totalGroups, totalEvents, totalMembers, totalParticipants, totalViewerLinks, dbPath)
	fmt.Printf("Login: admin@bandcash.local\n")
}
