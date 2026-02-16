package entry

import (
	"log/slog"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type entryInlineParams struct {
	FormData entryData `json:"formData"`
	Mode     string    `json:"mode"`
}

type modeParams struct {
	Mode string `json:"mode"`
}

type entryData struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Time        string `json:"time" validate:"required"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
}

type participantParams struct {
	ParticipantForm participantData `json:"participantForm"`
}

type participantData struct {
	PayeeID   int64  `json:"payeeId" validate:"required,gt=0"`
	PayeeName string `json:"payeeName"`
	Amount    int64  `json:"amount" validate:"required,gte=0"`
	Expense   int64  `json:"expense" validate:"gte=0"`
}

type participantTableParams struct {
	FormData participantData `json:"formData"`
}

// Default signal states for resetting forms on success
var (
	defaultEntrySignals = map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
	}
	defaultParticipantSignals = map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": 0, "payeeName": "", "amount": 0, "expense": 0},
	}
	// Error field lists for validation
	entryErrorFields       = []string{"title", "time", "description", "amount"}
	participantErrorFields = []string{"payeeId", "amount", "expense"}
)

func (e *Entries) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("entry.list: failed to get data", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("entry.index", "entry_count", len(data.Entries))
	return utils.RenderTemplate(c.Response().Writer, e.tmpl, "index", data)
}

func (e *Entries) Show(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("entry.show: invalid id", "err", err)
		return c.NoContent(400)
	}

	data, err := e.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("entry.show: failed to get data", "err", err)
		return c.NoContent(500)
	}

	return utils.RenderTemplate(c.Response().Writer, e.tmpl, "show", data)
}

func (e *Entries) Edit(c echo.Context) error {
	utils.EnsureClientID(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("entry.edit: invalid id", "err", err)
		return c.NoContent(400)
	}

	data, err := e.GetEditData(c.Request().Context(), id)
	if err != nil {
		slog.Error("entry.edit: failed to get data", "err", err)
		return c.NoContent(500)
	}

	return utils.RenderTemplate(c.Response().Writer, e.tmpl, "edit", data)
}

func (e *Entries) Create(c echo.Context) error {
	var signals entryInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("entry.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	slog.Debug("entry.create.table: signals received", "formData", signals.FormData)

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		slog.Debug("entry.create.table: validation failed", "errors", errs)
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	entry, err := db.Qry.CreateEntry(c.Request().Context(), db.CreateEntryParams{
		Title:       signals.FormData.Title,
		Time:        signals.FormData.Time,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
	})
	if err != nil {
		slog.Error("entry.create.table: failed to create entry", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("entry.create.table", "id", entry.ID, "title", entry.Title)

	utils.SSEHub.PatchSignals(c, defaultEntrySignals)
	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("entry.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}

	html, err := utils.RenderBlock(e.tmpl, "entry-index", data)
	if err != nil {
		slog.Error("entry.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Entries) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("entry.update: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals entryInlineParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("entry.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(entryErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = db.Qry.UpdateEntry(c.Request().Context(), db.UpdateEntryParams{
		Title:       signals.FormData.Title,
		Time:        signals.FormData.Time,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
		ID:          int64(id),
	})
	if err != nil {
		slog.Error("entry.update: failed to update entry", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("entry.update", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/entry/"+strconv.Itoa(id))
		if err != nil {
			slog.Warn("entry.update: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultEntrySignals)
	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("entry.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(e.tmpl, "entry-index", data)
	if err != nil {
		slog.Error("entry.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Entries) Destroy(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("entry.destroy: invalid id", "err", err)
		return c.NoContent(400)
	}

	var signals modeParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("entry.destroy: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	err = db.Qry.DeleteEntry(c.Request().Context(), int64(id))
	if err != nil {
		slog.Error("entry.destroy: failed to delete entry", "err", err)
		return c.NoContent(500)
	}

	slog.Debug("entry.destroy", "id", id)

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/entry")
		if err != nil {
			slog.Warn("entry.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultEntrySignals)
	data, err := e.GetIndexData(c.Request().Context())
	if err != nil {
		slog.Error("entry.destroy: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(e.tmpl, "entry-index", data)
	if err != nil {
		slog.Error("entry.destroy: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Entries) CreateParticipant(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("participant.create.table: invalid entry id", "err", err)
		return c.NoContent(400)
	}

	var signals participantTableParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("participant.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	// Set default expense to 0 if not provided
	expense := signals.FormData.Expense

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		EntryID: int64(id),
		PayeeID: signals.FormData.PayeeID,
		Amount:  signals.FormData.Amount,
		Expense: expense,
	})
	if err != nil {
		slog.Error("participant.create.table: failed to add participant", "err", err)
		return c.NoContent(500)
	}

	data, err := e.GetShowData(c.Request().Context(), id)
	if err != nil {
		slog.Error("participant.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(e.tmpl, "entry-show", data)
	if err != nil {
		slog.Error("participant.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.create.table", "entry_id", id, "payee_id", signals.FormData.PayeeID)
	return c.NoContent(200)
}

func (e *Entries) UpdateParticipant(c echo.Context) error {
	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("participant.update: invalid entry id", "err", err)
		return c.NoContent(400)
	}

	payeeID, err := strconv.Atoi(c.Param("payeeId"))
	if err != nil {
		slog.Warn("participant.update: invalid payee id", "err", err)
		return c.NoContent(400)
	}

	var signals participantTableParams
	err = datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("participant.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.Validate(signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	// Set default expense to 0 if not provided
	expense := signals.FormData.Expense

	err = db.Qry.UpdateParticipant(c.Request().Context(), db.UpdateParticipantParams{
		Amount:  signals.FormData.Amount,
		Expense: expense,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	})
	if err != nil {
		slog.Error("participant.update: failed to update participant", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)
	data, err := e.GetShowData(c.Request().Context(), entryID)
	if err != nil {
		slog.Error("participant.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(e.tmpl, "entry-show", data)
	if err != nil {
		slog.Error("participant.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	slog.Debug("participant.update", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) DeleteParticipantTable(c echo.Context) error {
	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("participant.delete: invalid entry id", "err", err)
		return c.NoContent(400)
	}

	payeeID, err := strconv.Atoi(c.Param("payeeId"))
	if err != nil {
		slog.Warn("participant.delete: invalid payee id", "err", err)
		return c.NoContent(400)
	}

	err = db.Qry.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	})
	if err != nil {
		slog.Error("participant.delete: failed to remove participant", "err", err)
		return c.NoContent(500)
	}

	data, err := e.GetShowData(c.Request().Context(), entryID)
	if err != nil {
		slog.Error("participant.delete: failed to get data", "err", err)
		return c.NoContent(500)
	}
	html, err := utils.RenderBlock(e.tmpl, "entry-show", data)
	if err != nil {
		slog.Error("participant.delete: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.delete", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}
