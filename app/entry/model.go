package entry

import (
	"database/sql"
	"fmt"
	"html/template"
	"time"

	"webapp/internal/db"
)

type Entry struct {
	ID          int
	Title       string
	Time        string
	Description string
	Amount      float64
	CreatedAt   time.Time
}

type EntryData struct {
	Title string
	Entry *Entry
}

type EntriesData struct {
	Title   string
	Entries []Entry
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

func (e *Entries) CreateEntry(title, entryTime, description string, amount float64) (*Entry, error) {
	result, err := db.DB.Exec(
		"INSERT INTO entries (title, time, description, amount) VALUES (?, ?, ?, ?)",
		title, entryTime, description, amount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &Entry{
		ID:          int(id),
		Title:       title,
		Time:        entryTime,
		Description: description,
		Amount:      amount,
	}, nil
}

func (e *Entries) GetEntry(id int) (*Entry, error) {
	row := db.DB.QueryRow(
		"SELECT id, title, time, description, amount, created_at FROM entries WHERE id = ?",
		id,
	)

	var entry Entry
	var createdAt sql.NullTime
	err := row.Scan(&entry.ID, &entry.Title, &entry.Time, &entry.Description, &entry.Amount, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan entry: %w", err)
	}
	if createdAt.Valid {
		entry.CreatedAt = createdAt.Time
	}

	return &entry, nil
}

func (e *Entries) UpdateEntry(id int, title, entryTime, description string, amount float64) error {
	_, err := db.DB.Exec(
		"UPDATE entries SET title = ?, time = ?, description = ?, amount = ? WHERE id = ?",
		title, entryTime, description, amount, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}
	return nil
}

func (e *Entries) AllEntries() ([]Entry, error) {
	rows, err := db.DB.Query(
		"SELECT id, title, time, description, amount, created_at FROM entries ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var createdAt sql.NullTime
		err := rows.Scan(&entry.ID, &entry.Title, &entry.Time, &entry.Description, &entry.Amount, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		if createdAt.Valid {
			entry.CreatedAt = createdAt.Time
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating entries: %w", err)
	}

	return entries, nil
}

func (e *Entries) DeleteEntry(id int) error {
	_, err := db.DB.Exec("DELETE FROM entries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

func (e *Entries) GetEntriesData() (any, error) {
	entries, err := e.AllEntries()
	if err != nil {
		return nil, err
	}
	return EntriesData{
		Title:   "Entries",
		Entries: entries,
	}, nil
}
