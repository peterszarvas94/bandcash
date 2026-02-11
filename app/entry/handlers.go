package entry

import (
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	appmw "webapp/internal/middleware"
	"webapp/internal/utils"
)

type createSignals struct {
	Title       string  `json:"title"`
	Time        string  `json:"time"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

func (e *Entries) List(c echo.Context) error {
	utils.EnsureClientID(c)

	log := appmw.Logger(c)
	data, err := e.GetEntriesData()
	if err != nil {
		log.Error("entry.list: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.list", "entry_count", len(data.(EntriesData).Entries))
	return e.tmpl.ExecuteTemplate(c.Response().Writer, "list", data)
}

func (e *Entries) New(c echo.Context) error {
	utils.EnsureClientID(c)
	return e.tmpl.ExecuteTemplate(c.Response().Writer, "new", EntryData{Title: "New Entry"})
}

func (e *Entries) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	entry, err := e.GetEntry(id)
	if err != nil {
		log.Error("entry.show: failed to get entry", "err", err)
		return c.String(500, "Internal Server Error")
	}
	if entry == nil {
		return c.String(404, "Entry not found")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "show", EntryData{Title: entry.Title, Entry: entry})
}

func (e *Entries) Edit(c echo.Context) error {
	utils.EnsureClientID(c)
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	entry, err := e.GetEntry(id)
	if err != nil {
		log.Error("entry.edit: failed to get entry", "err", err)
		return c.String(500, "Internal Server Error")
	}
	if entry == nil {
		return c.String(404, "Entry not found")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "edit", EntryData{Title: "Edit Entry", Entry: entry})
}

func (e *Entries) Create(c echo.Context) error {
	log := appmw.Logger(c)

	var signals createSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.Title == "" {
		log.Debug("entry.create: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(signals.Title, signals.Time, signals.Description, signals.Amount)
	if err != nil {
		log.Error("entry.create: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.create", "id", entry.ID, "title", entry.Title)
	utils.Store.SignalAll()

	return c.NoContent(200)
}

func (e *Entries) Update(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals createSignals
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if err := e.UpdateEntry(id, signals.Title, signals.Time, signals.Description, signals.Amount); err != nil {
		log.Error("entry.update: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.update", "id", id)
	utils.Store.SignalAll()

	return c.NoContent(200)
}

func (e *Entries) Delete(c echo.Context) error {
	log := appmw.Logger(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("entry.delete: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := e.DeleteEntry(id); err != nil {
		log.Error("entry.delete: failed to delete entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.delete", "id", id)
	utils.Store.SignalAll()

	return c.NoContent(200)
}

func (e *Entries) DataForSSE() any {
	data, err := e.GetEntriesData()
	if err != nil {
		slog.Error("entry.DataForSSE: failed to get data", "err", err)
		return EntriesData{Title: "Entries", Entries: []Entry{}}
	}
	return data
}
