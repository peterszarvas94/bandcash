package entry

import (
	"context"
	"html/template"
	"log/slog"
	"strconv"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EntryData struct {
	Title        string
	Entry        *db.Entry
	Participants []db.ListParticipantsByEntryRow
	Payees       []db.Payee
	PayeeIDs     map[int64]bool
	Breadcrumbs  []utils.Crumb
}

type EntriesData struct {
	Title       string
	Entries     []db.Entry
	Breadcrumbs []utils.Crumb
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

func (e *Entries) GetShowData(ctx context.Context, id int) (EntryData, error) {
	entry, err := db.Qry.GetEntry(ctx, int64(id))
	if err != nil {
		return EntryData{}, err
	}

	participants, err := db.Qry.ListParticipantsByEntry(ctx, int64(id))
	if err != nil {
		return EntryData{}, err
	}

	payees, err := db.Qry.ListPayees(ctx)
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
		Entry:        &entry,
		Participants: participants,
		Payees:       filteredPayees,
		PayeeIDs:     payeeIDs,
		Breadcrumbs: []utils.Crumb{
			{Label: "Entries", Href: "/entry"},
			{Label: entry.Title},
		},
	}, nil
}

func (e *Entries) GetEditData(ctx context.Context, id int) (EntryData, error) {
	entry, err := db.Qry.GetEntry(ctx, int64(id))
	if err != nil {
		return EntryData{}, err
	}

	return EntryData{
		Title: "Edit Entry",
		Entry: &entry,
		Breadcrumbs: []utils.Crumb{
			{Label: "Entries", Href: "/entry"},
			{Label: entry.Title, Href: "/entry/" + strconv.Itoa(id)},
			{Label: "Edit"},
		},
	}, nil
}

func (e *Entries) GetIndexData(ctx context.Context) (any, error) {
	entries, err := db.Qry.ListEntries(ctx)
	if err != nil {
		return nil, err
	}

	return EntriesData{
		Title:   "Entries",
		Entries: entries,
		Breadcrumbs: []utils.Crumb{
			{Label: "Entries"},
		},
	}, nil
}
