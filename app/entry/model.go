package entry

import (
	"context"
	"html/template"
	"log/slog"

	"bandcash/internal/db"
)

type EntryData struct {
	Title        string
	Entry        *db.Entry
	Participants []db.ListParticipantsByEntryRow
	Payees       []db.Payee
	PayeeIDs     map[int64]bool
}

type EntriesData struct {
	Title   string
	Entries []db.Entry
}

type Entries struct {
	tmpl *template.Template
}

func New() *Entries {
	return &Entries{}
}

func (e *Entries) SetTemplate(tmpl *template.Template) {
	e.tmpl = tmpl
}

func (e *Entries) CreateEntry(ctx context.Context, title, entryTime, description string, amount float64) (*db.Entry, error) {
	entry, err := db.Qry.CreateEntry(ctx, db.CreateEntryParams{
		Title:       title,
		Time:        entryTime,
		Description: description,
		Amount:      amount,
	})
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (e *Entries) GetEntry(ctx context.Context, id int) (*db.Entry, error) {
	entry, err := db.Qry.GetEntry(ctx, int64(id))
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (e *Entries) GetParticipants(ctx context.Context, entryID int) ([]db.ListParticipantsByEntryRow, error) {
	return db.Qry.ListParticipantsByEntry(ctx, int64(entryID))
}

func (e *Entries) GetPayees(ctx context.Context) ([]db.Payee, error) {
	return db.Qry.ListPayees(ctx)
}

func (e *Entries) GetShowData(ctx context.Context, id int) (EntryData, error) {
	entry, err := e.GetEntry(ctx, id)
	if err != nil {
		return EntryData{}, err
	}

	participants, err := e.GetParticipants(ctx, id)
	if err != nil {
		return EntryData{}, err
	}

	payees, err := e.GetPayees(ctx)
	if err != nil {
		return EntryData{}, err
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

	slog.Info("entry.show.data", "entry_id", id, "participants", len(participants), "payees_total", len(payees), "payees_filtered", len(filteredPayees))

	return EntryData{
		Title:        entry.Title,
		Entry:        entry,
		Participants: participants,
		Payees:       filteredPayees,
		PayeeIDs:     payeeIDs,
	}, nil
}

func (e *Entries) UpdateEntry(ctx context.Context, id int, title, entryTime, description string, amount float64) (*db.Entry, error) {
	updated, err := db.Qry.UpdateEntry(ctx, db.UpdateEntryParams{
		Title:       title,
		Time:        entryTime,
		Description: description,
		Amount:      amount,
		ID:          int64(id),
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (e *Entries) AllEntries(ctx context.Context) ([]db.Entry, error) {
	return db.Qry.ListEntries(ctx)
}

func (e *Entries) DeleteEntry(ctx context.Context, id int) error {
	return db.Qry.DeleteEntry(ctx, int64(id))
}

func (e *Entries) GetIndexData(ctx context.Context) (any, error) {
	entries, err := e.AllEntries(ctx)
	if err != nil {
		return nil, err
	}
	return EntriesData{
		Title:   "Entries",
		Entries: entries,
	}, nil
}
