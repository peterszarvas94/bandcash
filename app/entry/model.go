package entry

import (
	"context"
	"html/template"

	"bandcash/internal/db"
)

type EntryData struct {
	Title string
	Entry *db.Entry
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

func (e *Entries) GetEntriesData(ctx context.Context) (any, error) {
	entries, err := e.AllEntries(ctx)
	if err != nil {
		return nil, err
	}
	return EntriesData{
		Title:   "Entries",
		Entries: entries,
	}, nil
}
