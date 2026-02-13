package entry

import (
	"encoding/json"
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/hub"
	"bandcash/internal/utils"
	"bandcash/internal/view"
)

type entryParams struct {
	Title       string          `json:"title"`
	Time        string          `json:"time"`
	Description string          `json:"description"`
	Amount      json.RawMessage `json:"amount"`
}

type entryInlineParams struct {
	FormData entryData `json:"formData"`
}

type entryData struct {
	Title       string          `json:"title"`
	Time        string          `json:"time"`
	Description string          `json:"description"`
	Amount      json.RawMessage `json:"amount"`
}

type participantParams struct {
	ParticipantForm participantData `json:"participantForm"`
}

type participantData struct {
	PayeeID   json.RawMessage `json:"payeeId"`
	PayeeName string          `json:"payeeName"`
	Amount    json.RawMessage `json:"amount"`
}

type participantTableParams struct {
	FormData participantData `json:"formData"`
}

// Default signal states for resetting forms
var (
	defaultEntrySignals = map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
	}
	defaultParticipantSignals = map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}
)

func (e *Entries) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("entry.list: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.index", "entry_count", len(data.(EntriesData).Entries))
	return e.tmpl.ExecuteTemplate(c.Response().Writer, "index", data)
}

func (e *Entries) New(c echo.Context) error {
	utils.EnsureClientID(c)
	return e.tmpl.ExecuteTemplate(c.Response().Writer, "new", EntryData{
		Title: "New Entry",
		Breadcrumbs: []view.Crumb{
			{Label: "Entries", Href: "/entry"},
			{Label: "New"},
		},
	})
}

func (e *Entries) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	data, err := e.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("entry.show: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "show", data)
}

func (e *Entries) Edit(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	entry, err := e.GetEntry(c.Request().Context(), id)
	if err != nil {
		slog.Error("entry.edit: failed to get entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "edit", EntryData{
		Title: "Edit Entry",
		Entry: entry,
		Breadcrumbs: []view.Crumb{
			{Label: "Entries", Href: "/entry"},
			{Label: entry.Title, Href: "/entry/" + strconv.Itoa(id)},
			{Label: "Edit"},
		},
	})
}

func (e *Entries) Create(c echo.Context) error {
	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("entry.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("entry.create: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	if signals.FormData.Title == "" {
		slog.Debug("entry.create: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(c.Request().Context(), signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.create: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.create", "id", entry.ID, "title", entry.Title)

	if err := hub.Hub.Redirect(c, "/entry/"+strconv.FormatInt(entry.ID, 10)); err != nil {
		slog.Warn("entry.create: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) CreateTable(c echo.Context) error {

	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("entry.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("entry.create.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	if signals.FormData.Title == "" {
		slog.Debug("entry.create.table: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(c.Request().Context(), signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.create.table: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.create.table", "id", entry.ID, "title", entry.Title)

	if err := hub.Hub.PatchSignals(c, defaultEntrySignals); err != nil {
		slog.Warn("entry.create.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("entry.create.table: failed to signal client", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) Update(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("entry.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("entry.update: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = e.UpdateEntry(c.Request().Context(), id, signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.update: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.update", "id", id)

	if err := hub.Hub.Redirect(c, "/entry/"+strconv.Itoa(id)); err != nil {
		slog.Warn("entry.update: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) UpdateTable(c echo.Context) error {

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("entry.update.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("entry.update.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = e.UpdateEntry(c.Request().Context(), id, signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.update.table: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.update.table", "id", id)

	if err := hub.Hub.PatchSignals(c, defaultEntrySignals); err != nil {
		slog.Warn("entry.update.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("entry.update.table: failed to signal client", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) DestroyTable(c echo.Context) error {

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("entry.destroy: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := e.DeleteEntry(c.Request().Context(), id); err != nil {
		slog.Error("entry.destroy: failed to delete entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.destroy", "id", id)

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("entry.destroy: failed to signal client", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) Destroy(c echo.Context) error {

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("entry.destroy: invalid id", "id", idStr)
		return c.NoContent(400)
	}

	if err := e.DeleteEntry(c.Request().Context(), id); err != nil {
		slog.Error("entry.destroy: failed to delete entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.destroy", "id", id)

	if err := hub.Hub.Redirect(c, "/entry"); err != nil {
		slog.Warn("entry.destroy: failed to redirect", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) AddParticipant(c echo.Context) error {

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals participantParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("participant.create: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	payeeID, err := utils.ParseRawInt64(signals.ParticipantForm.PayeeID)
	if err != nil {
		slog.Warn("participant.create: invalid payee_id", "payee_id", string(signals.ParticipantForm.PayeeID))
		return c.String(400, "Invalid payee")
	}

	amount, err := utils.ParseRawInt64(signals.ParticipantForm.Amount)
	if err != nil {
		slog.Warn("participant.create: invalid amount", "amount", string(signals.ParticipantForm.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		EntryID: int64(id),
		PayeeID: payeeID,
		Amount:  amount,
	})
	if err != nil {
		slog.Error("participant.create: failed to add participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("participant.create: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(c, defaultParticipantSignals); err != nil {
		slog.Warn("participant.create: failed to patch signals", "err", err)
	}

	slog.Debug("participant.create", "entry_id", id, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) AddParticipantTable(c echo.Context) error {

	id, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals participantTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("participant.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	payeeID, err := utils.ParseRawInt64(signals.FormData.PayeeID)
	if err != nil {
		slog.Warn("participant.create.table: invalid payee_id", "payee_id", string(signals.FormData.PayeeID))
		return c.String(400, "Invalid payee")
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("participant.create.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		EntryID: int64(id),
		PayeeID: payeeID,
		Amount:  amount,
	})
	if err != nil {
		slog.Error("participant.create.table: failed to add participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("participant.create.table: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(c, defaultParticipantSignals); err != nil {
		slog.Warn("participant.create.table: failed to patch signals", "err", err)
	}

	slog.Debug("participant.create.table", "entry_id", id, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) UpdateParticipant(c echo.Context) error {

	entryID, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := utils.ParamInt(c, "payeeId")
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	var signals participantParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("participant.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.ParticipantForm.Amount)
	if err != nil {
		slog.Warn("participant.update: invalid amount", "amount", string(signals.ParticipantForm.Amount))
		return c.String(400, "Invalid amount")
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		slog.Error("participant.update: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("participant.update: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(c, defaultParticipantSignals); err != nil {
		slog.Warn("participant.update: failed to patch signals", "err", err)
	}

	slog.Debug("participant.update", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) UpdateParticipantTable(c echo.Context) error {

	entryID, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := utils.ParamInt(c, "payeeId")
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	var signals participantTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Warn("participant.update.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		slog.Warn("participant.update.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		slog.Error("participant.update.table: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("participant.update.table: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(c, defaultParticipantSignals); err != nil {
		slog.Warn("participant.update.table: failed to patch signals", "err", err)
	}

	slog.Debug("participant.update.table", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) DeleteParticipantTable(c echo.Context) error {

	entryID, err := utils.ParamInt(c, "id")
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := utils.ParamInt(c, "payeeId")
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	if err := db.Qry.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		slog.Error("participant.delete: failed to remove participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	if err := hub.Hub.Refresh(c); err != nil {
		slog.Warn("participant.delete: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(c, defaultParticipantSignals); err != nil {
		slog.Warn("participant.delete: failed to patch signals", "err", err)
	}

	slog.Debug("participant.delete", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}
