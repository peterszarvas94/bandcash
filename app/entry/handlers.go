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
	"bandcash/internal/validation"
	"bandcash/internal/view"
)

type entryInlineParams struct {
	FormData entryData `json:"formData"`
}

type entryData struct {
	Title       string          `json:"title" validate:"required,min=1,max=255"`
	Time        string          `json:"time" validate:"required"`
	Description string          `json:"description" validate:"max=1000"`
	Amount      json.RawMessage `json:"amount" validate:"required"`
}

type participantParams struct {
	ParticipantForm participantData `json:"participantForm"`
}

type participantData struct {
	PayeeID   json.RawMessage `json:"payeeId" validate:"required"`
	PayeeName string          `json:"payeeName"`
	Amount    json.RawMessage `json:"amount" validate:"required"`
}

type participantTableParams struct {
	FormData participantData `json:"formData"`
}

// Default signal states for resetting forms on success
var (
	defaultEntrySignals = map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
	}
	defaultParticipantSignals = map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": 0, "payeeName": "", "amount": 0},
	}
	// Error field lists for validation
	entryErrorFields       = []string{"title", "time", "description", "amount"}
	participantErrorFields = []string{"payeeId", "amount"}
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

	slog.Debug("entry.create: signals received", "formData", signals.FormData)

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		slog.Debug("entry.create: validation failed", "errors", errs)
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
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

	slog.Debug("entry.create.table: signals received", "formData", signals.FormData)

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		slog.Debug("entry.create.table: validation failed", "errors", errs)
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
	}

	entry, err := e.CreateEntry(c.Request().Context(), signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.create.table: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.create.table", "id", entry.ID, "title", entry.Title)

	hub.Hub.PatchSignals(c, defaultEntrySignals)
	hub.Hub.Refresh(c)

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

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
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

	slog.Debug("entry.update.table: signals received", "formData", signals.FormData)

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		slog.Debug("entry.update.table: validation failed", "errors", errs)
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(entryErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
	}

	_, err = e.UpdateEntry(c.Request().Context(), id, signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		slog.Error("entry.update.table: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.update.table", "id", id)

	hub.Hub.PatchSignals(c, defaultEntrySignals)
	hub.Hub.Refresh(c)

	return c.NoContent(200)
}

func (e *Entries) DestroyTable(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		slog.Warn("entry.destroy.table: invalid id", "err", err)
		return c.NoContent(400)
	}

	if err := e.DeleteEntry(c.Request().Context(), id); err != nil {
		slog.Error("entry.destroy.table: failed to delete entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	slog.Debug("entry.destroy.table", "id", id)

	hub.Hub.Refresh(c)

	return c.NoContent(200)
}

func (e *Entries) Destroy(c echo.Context) error {
	id, err := utils.ParamInt(c, "id")
	if err != nil {
		slog.Warn("entry.destroy: invalid id", "err", err)
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

	// Validate
	if errs := validation.ValidateStruct(signals.ParticipantForm); errs != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	payeeID, err := strconv.ParseInt(string(signals.ParticipantForm.PayeeID), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"payeeId": "Invalid payee"})})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.ParticipantForm.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
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

	hub.Hub.PatchSignals(c, defaultParticipantSignals)
	hub.Hub.Refresh(c)

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

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	payeeID, err := strconv.ParseInt(string(signals.FormData.PayeeID), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"payeeId": "Invalid payee"})})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
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

	hub.Hub.Refresh(c)
	hub.Hub.PatchSignals(c, defaultParticipantSignals)

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

	// Validate
	if errs := validation.ValidateStruct(signals.ParticipantForm); errs != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.ParticipantForm.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		slog.Error("participant.update: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	hub.Hub.PatchSignals(c, defaultParticipantSignals)
	hub.Hub.Refresh(c)

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

	// Validate
	if errs := validation.ValidateStruct(signals.FormData); errs != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	amount, err := strconv.ParseInt(string(signals.FormData.Amount), 10, 64)
	if err != nil {
		hub.Hub.PatchSignals(c, map[string]any{"errors": validation.WithErrors(participantErrorFields, map[string]string{"amount": "Invalid amount"})})
		return c.NoContent(422)
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		slog.Error("participant.update.table: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	hub.Hub.Refresh(c)
	hub.Hub.PatchSignals(c, defaultParticipantSignals)

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

	hub.Hub.Refresh(c)
	hub.Hub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.delete", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}
