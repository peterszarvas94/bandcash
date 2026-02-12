package entry

import (
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	appmw "bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type entryParams struct {
	Title       string  `json:"title"`
	Time        string  `json:"time"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type participantParams struct {
	PayeeID json.RawMessage `json:"payee_id"`
	Amount  json.RawMessage `json:"amount"`
}

func (e *Entries) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	log := appmw.Logger(c)
	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		log.Error("entry.list: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.index", "entry_count", len(data.(EntriesData).Entries))
	return e.tmpl.ExecuteTemplate(c.Response().Writer, "index", data)
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

	entry, err := e.GetEntry(c.Request().Context(), id)
	if err != nil {
		log.Error("entry.show: failed to get entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	participants, err := e.GetParticipants(c.Request().Context(), id)
	if err != nil {
		log.Error("entry.show: failed to get participants", "err", err)
		return c.String(500, "Internal Server Error")
	}

	payees, err := e.GetPayees(c.Request().Context())
	if err != nil {
		log.Error("entry.show: failed to get payees", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "show", EntryData{
		Title:        entry.Title,
		Entry:        entry,
		Participants: participants,
		Payees:       payees,
	})
}

func (e *Entries) Edit(c echo.Context) error {
	utils.EnsureClientID(c)
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	entry, err := e.GetEntry(c.Request().Context(), id)
	if err != nil {
		log.Error("entry.edit: failed to get entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "edit", EntryData{Title: "Edit Entry", Entry: entry})
}

func (e *Entries) Create(c echo.Context) error {
	log := appmw.Logger(c)

	var signals entryParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if signals.Title == "" {
		log.Debug("entry.create: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(c.Request().Context(), signals.Title, signals.Time, signals.Description, signals.Amount)
	if err != nil {
		log.Error("entry.create: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.create", "id", entry.ID, "title", entry.Title)
	return c.Redirect(303, "/entry")
}

func (e *Entries) Update(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals entryParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	_, err = e.UpdateEntry(c.Request().Context(), id, signals.Title, signals.Time, signals.Description, signals.Amount)
	if err != nil {
		log.Error("entry.update: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.update", "id", id)
	return c.Redirect(303, "/entry/"+strconv.Itoa(id))
}

func (e *Entries) Destroy(c echo.Context) error {
	log := appmw.Logger(c)

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Warn("entry.destroy: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := e.DeleteEntry(c.Request().Context(), id); err != nil {
		log.Error("entry.destroy: failed to delete entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.destroy", "id", id)
	return c.Redirect(303, "/entry")
}

func (e *Entries) AddParticipant(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals participantParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("participant.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	if len(signals.PayeeID) == 0 || string(signals.PayeeID) == "\"\"" {
		log.Debug("participant.create: empty payee_id")
		return c.Redirect(303, "/entry/"+strconv.Itoa(id))
	}

	payeeID, err := utils.ParseRawInt64(signals.PayeeID)
	if err != nil {
		log.Warn("participant.create: invalid payee_id", "payee_id", string(signals.PayeeID))
		return c.String(400, "Invalid payee")
	}

	amount, err := utils.ParseRawFloat64(signals.Amount)
	if err != nil {
		log.Warn("participant.create: invalid amount", "amount", string(signals.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		EntryID: int64(id),
		PayeeID: payeeID,
		Amount:  amount,
	})
	if err != nil {
		log.Error("participant.create: failed to add participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("participant.create", "entry_id", id, "payee_id", signals.PayeeID)
	return c.Redirect(303, "/entry/"+strconv.Itoa(id))
}
