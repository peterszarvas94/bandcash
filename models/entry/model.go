package entry

import (
	"context"
	"log/slog"

	"bandcash/internal/db"
	"bandcash/internal/utils"
	entryview "bandcash/models/entry/templates/view"
)

type Entries struct {
}

func New() *Entries {
	return &Entries{}
}

func (e *Entries) GetShowData(ctx context.Context, id int) (entryview.EntryData, error) {
	entry, err := db.Qry.GetEntry(ctx, int64(id))
	if err != nil {
		return entryview.EntryData{}, err
	}

	participants, err := db.Qry.ListParticipantsByEntry(ctx, int64(id))
	if err != nil {
		return entryview.EntryData{}, err
	}

	payees, err := db.Qry.ListPayees(ctx)
	if err != nil {
		return entryview.EntryData{}, err
	}

	payeeIDs := make(map[int64]bool, len(participants))
	for _, participant := range participants {
		payeeIDs[participant.ID] = true
	}

	filteredPayees := make([]db.Payee, 0, len(payees))
	for _, payee := range payees {
		if payeeIDs[payee.ID] {
			continue
		}
		filteredPayees = append(filteredPayees, payee)
	}

	// Calculate total distributed and leftover
	var totalDistributed int64
	for _, p := range participants {
		totalDistributed += p.ParticipantAmount + p.ParticipantExpense
	}
	leftover := entry.Amount - totalDistributed

	slog.Info("entry.show.data", "entry_id", id, "participants", len(participants), "payees_total", len(payees), "payees_filtered", len(filteredPayees), "leftover", leftover)

	return entryview.EntryData{
		Title:            entry.Title,
		Entry:            &entry,
		Participants:     participants,
		Payees:           filteredPayees,
		PayeeIDs:         payeeIDs,
		Leftover:         leftover,
		TotalDistributed: totalDistributed,
		Breadcrumbs: []utils.Crumb{
			{Label: "Entries", Href: "/entry"},
			{Label: entry.Title},
		},
	}, nil
}

func (e *Entries) GetIndexData(ctx context.Context) (entryview.EntriesData, error) {
	entries, err := db.Qry.ListEntries(ctx)
	if err != nil {
		return entryview.EntriesData{}, err
	}

	return entryview.EntriesData{
		Title:   "Entries",
		Entries: entries,
		Breadcrumbs: []utils.Crumb{
			{Label: "Entries"},
		},
	}, nil
}
