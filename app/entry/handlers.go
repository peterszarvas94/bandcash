package entry

import (
	"encoding/json"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/hub"
	appmw "bandcash/internal/middleware"
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
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	data, err := e.GetShowData(c.Request().Context(), id)
	if err != nil {
		log.Error("entry.show: failed to get data", "err", err)
		return c.String(500, "Internal Server Error")
	}

	return e.tmpl.ExecuteTemplate(c.Response().Writer, "show", data)
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
	log := appmw.Logger(c)
	var title, entryTime, description string
	var amount int64

	var signals entryParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		signals.Title = c.FormValue("title")
		signals.Time = c.FormValue("time")
		signals.Description = c.FormValue("description")
		amountVal := c.FormValue("amount")
		if amountVal != "" {
			signals.Amount = json.RawMessage(amountVal)
		}
	}

	var err error
	amount, err = utils.ParseRawInt64(signals.Amount)
	if err != nil {
		log.Warn("entry.create: invalid amount", "amount", string(signals.Amount))
		return c.String(400, "Invalid amount")
	}
	title = signals.Title
	entryTime = signals.Time
	description = signals.Description

	if title == "" {
		log.Debug("entry.create: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(c.Request().Context(), title, entryTime, description, amount)
	if err != nil {
		log.Error("entry.create: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.create", "id", entry.ID, "title", entry.Title)
	return c.Redirect(303, "/entry")
}

func (e *Entries) CreateTable(c echo.Context) error {
	log := appmw.Logger(c)

	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		log.Warn("entry.create.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	if signals.FormData.Title == "" {
		log.Debug("entry.create.table: empty title")
		return c.NoContent(200)
	}

	entry, err := e.CreateEntry(c.Request().Context(), signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		log.Error("entry.create.table: failed to create entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.create.table", "id", entry.ID, "title", entry.Title)

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("entry.create.table: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
	}); err != nil {
		log.Warn("entry.create.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("entry.create.table: failed to signal client", "err", err)
	}

	return c.NoContent(200)
}

func (e *Entries) Update(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var title, entryTime, description string
	var amount int64

	var signals entryParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		signals.Title = c.FormValue("title")
		signals.Time = c.FormValue("time")
		signals.Description = c.FormValue("description")
		amountVal := c.FormValue("amount")
		if amountVal != "" {
			signals.Amount = json.RawMessage(amountVal)
		}
	}

	amount, err = utils.ParseRawInt64(signals.Amount)
	if err != nil {
		log.Warn("entry.update: invalid amount", "amount", string(signals.Amount))
		return c.String(400, "Invalid amount")
	}
	title = signals.Title
	entryTime = signals.Time
	description = signals.Description

	_, err = e.UpdateEntry(c.Request().Context(), id, title, entryTime, description, amount)
	if err != nil {
		log.Error("entry.update: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.update", "id", id)
	return c.Redirect(303, "/entry/"+strconv.Itoa(id))
}

func (e *Entries) UpdateTable(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals entryInlineParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("entry.update.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		log.Warn("entry.update.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = e.UpdateEntry(c.Request().Context(), id, signals.FormData.Title, signals.FormData.Time, signals.FormData.Description, amount)
	if err != nil {
		log.Error("entry.update.table: failed to update entry", "err", err)
		return c.String(500, "Internal Server Error")
	}

	log.Debug("entry.update.table", "id", id)

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("entry.update.table: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
	}); err != nil {
		log.Warn("entry.update.table: failed to patch signals", "err", err)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("entry.update.table: failed to signal client", "err", err)
	}

	return c.NoContent(200)
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

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("entry.destroy: failed to signal client", "err", err)
	}

	return c.NoContent(200)
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

	payeeID, err := utils.ParseRawInt64(signals.ParticipantForm.PayeeID)
	if err != nil {
		log.Warn("participant.create: invalid payee_id", "payee_id", string(signals.ParticipantForm.PayeeID))
		return c.String(400, "Invalid payee")
	}

	amount, err := utils.ParseRawInt64(signals.ParticipantForm.Amount)
	if err != nil {
		log.Warn("participant.create: invalid amount", "amount", string(signals.ParticipantForm.Amount))
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

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("participant.create: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("participant.create: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}); err != nil {
		log.Warn("participant.create: failed to patch signals", "err", err)
	}

	log.Debug("participant.create", "entry_id", id, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) AddParticipantTable(c echo.Context) error {
	log := appmw.Logger(c)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid ID")
	}

	var signals participantTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("participant.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	payeeID, err := utils.ParseRawInt64(signals.FormData.PayeeID)
	if err != nil {
		log.Warn("participant.create.table: invalid payee_id", "payee_id", string(signals.FormData.PayeeID))
		return c.String(400, "Invalid payee")
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		log.Warn("participant.create.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		EntryID: int64(id),
		PayeeID: payeeID,
		Amount:  amount,
	})
	if err != nil {
		log.Error("participant.create.table: failed to add participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("participant.create.table: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("participant.create.table: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}); err != nil {
		log.Warn("participant.create.table: failed to patch signals", "err", err)
	}

	log.Debug("participant.create.table", "entry_id", id, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) UpdateParticipant(c echo.Context) error {
	log := appmw.Logger(c)

	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := strconv.Atoi(c.Param("payeeId"))
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	var signals participantParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("participant.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.ParticipantForm.Amount)
	if err != nil {
		log.Warn("participant.update: invalid amount", "amount", string(signals.ParticipantForm.Amount))
		return c.String(400, "Invalid amount")
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		log.Error("participant.update: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("participant.update: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("participant.update: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}); err != nil {
		log.Warn("participant.update: failed to patch signals", "err", err)
	}

	log.Debug("participant.update", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) UpdateParticipantTable(c echo.Context) error {
	log := appmw.Logger(c)

	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := strconv.Atoi(c.Param("payeeId"))
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	var signals participantTableParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		log.Warn("participant.update.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	amount, err := utils.ParseRawInt64(signals.FormData.Amount)
	if err != nil {
		log.Warn("participant.update.table: invalid amount", "amount", string(signals.FormData.Amount))
		return c.String(400, "Invalid amount")
	}

	if err := db.Qry.UpdateParticipantAmount(c.Request().Context(), db.UpdateParticipantAmountParams{
		Amount:  amount,
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		log.Error("participant.update.table: failed to update participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("participant.update.table: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("participant.update.table: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}); err != nil {
		log.Warn("participant.update.table: failed to patch signals", "err", err)
	}

	log.Debug("participant.update.table", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}

func (e *Entries) DeleteParticipant(c echo.Context) error {
	log := appmw.Logger(c)

	entryID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.String(400, "Invalid entry ID")
	}

	payeeID, err := strconv.Atoi(c.Param("payeeId"))
	if err != nil {
		return c.String(400, "Invalid payee ID")
	}

	if err := db.Qry.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
		EntryID: int64(entryID),
		PayeeID: int64(payeeID),
	}); err != nil {
		log.Error("participant.delete: failed to remove participant", "err", err)
		return c.String(500, "Internal Server Error")
	}

	clientID, err := utils.GetClientID(c)
	if err != nil {
		log.Warn("participant.delete: failed to read client_id", "err", err)
		return c.NoContent(200)
	}

	if err := hub.Hub.Render(c); err != nil {
		log.Warn("participant.delete: failed to signal client", "err", err)
	}

	if err := hub.Hub.PatchSignals(clientID, map[string]any{
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"payeeId": "", "payeeName": "", "amount": 0},
	}); err != nil {
		log.Warn("participant.delete: failed to patch signals", "err", err)
	}

	log.Debug("participant.delete", "entry_id", entryID, "payee_id", payeeID)
	return c.NoContent(200)
}
