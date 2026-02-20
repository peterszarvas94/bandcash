package event

import (
	"log/slog"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type eventInlineParams struct {
	FormData      eventData `json:"formData"`
	EventFormData eventData `json:"eventFormData"`
	Mode          string    `json:"mode"`
}

type modeParams struct {
	Mode string `json:"mode"`
}

type eventData struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Time        string `json:"time" validate:"required"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
}

type participantData struct {
	MemberID   string `json:"memberId" validate:"required"`
	MemberName string `json:"memberName"`
	Amount     int64  `json:"amount" validate:"required,gte=0"`
	Expense    int64  `json:"expense" validate:"gte=0"`
}

type participantTableParams struct {
	FormData participantData `json:"formData"`
}

// Default signal states for resetting forms on success
var (
	defaultEventSignals = map[string]any{
		"mode":      "",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"title": "", "time": "", "description": "", "amount": 0},
		"errors":    map[string]any{"title": "", "time": "", "description": "", "amount": ""},
	}
	defaultParticipantSignals = map[string]any{
		"formState":   "",
		"editingId":   "",
		"calcPercent": 0,
		"formData":    map[string]any{"memberId": "", "memberName": "", "amount": 0, "expense": 0},
		"errors":      map[string]any{"memberId": "", "amount": "", "expense": ""},
	}
	// Error field lists for validation
	eventErrorFields       = []string{"title", "time", "description", "amount"}
	participantErrorFields = []string{"memberId", "amount", "expense"}
)

func getGroupID(c echo.Context) string {
	return middleware.GetGroupID(c)
}

func getUserEmail(c echo.Context) string {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return ""
	}
	user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return ""
	}
	return user.Email
}

func (e *Events) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	data, err := e.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.list: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	slog.Debug("event.index", "event_count", len(data.Events))
	return utils.RenderComponent(c, EventIndex(data))
}

func (e *Events) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if id == "" {
		slog.Warn("event.show: invalid id")
		return c.NoContent(400)
	}

	data, err := e.GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("event.show: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	return utils.RenderComponent(c, EventShow(data))
}

func (e *Events) Create(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("event.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	slog.Debug("event.create.table: signals received", "formData", signals.FormData)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		slog.Debug("event.create.table: validation failed", "errors", errs)
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs)})
		return c.NoContent(422)
	}

	event, err := db.Qry.CreateEvent(c.Request().Context(), db.CreateEventParams{
		ID:          utils.GenerateID(utils.PrefixEvent),
		GroupID:     groupID,
		Title:       signals.FormData.Title,
		Time:        signals.FormData.Time,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
	})
	if err != nil {
		slog.Error("event.create.table: failed to create event", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "events.notifications.create_failed"))
		return c.NoContent(500)
	}

	slog.Debug("event.create.table", "id", event.ID, "title", event.Title)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "events.notifications.created"))

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	data, err := e.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	html, err := utils.RenderComponentStringFor(c, EventIndex(data))
	if err != nil {
		slog.Error("event.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Events) Update(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if id == "" {
		slog.Warn("event.update: invalid id")
		return c.NoContent(400)
	}

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("event.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	eventForm := signals.FormData
	if signals.EventFormData.Title != "" || signals.EventFormData.Time != "" || signals.EventFormData.Amount != 0 {
		eventForm = signals.EventFormData
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), eventForm); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs)})
		return c.NoContent(422)
	}

	_, err = db.Qry.UpdateEvent(c.Request().Context(), db.UpdateEventParams{
		Title:       eventForm.Title,
		Time:        eventForm.Time,
		Description: eventForm.Description,
		Amount:      eventForm.Amount,
		ID:          id,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("event.update: failed to update event", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(500)
	}

	slog.Debug("event.update", "id", id)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "events.notifications.updated"))

	if signals.Mode == "single" {
		utils.SSEHub.PatchSignals(c, map[string]any{
			"eventFormState": "",
			"eventFormData": map[string]any{
				"title":       eventForm.Title,
				"time":        eventForm.Time,
				"description": eventForm.Description,
				"amount":      eventForm.Amount,
			},
			"errors": map[string]any{"title": "", "time": "", "description": "", "amount": ""},
		})
		data, err := e.GetShowData(c.Request().Context(), groupID, id)
		if err != nil {
			slog.Error("event.update: failed to get data", "err", err)
			return c.NoContent(500)
		}
		data.IsAdmin = middleware.IsAdmin(c)
		data.UserEmail = userEmail
		html, err := utils.RenderComponentStringFor(c, EventShow(data))
		if err != nil {
			slog.Error("event.update: failed to render", "err", err)
			return c.NoContent(500)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	data, err := e.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, EventIndex(data))
	if err != nil {
		slog.Error("event.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Events) Destroy(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if id == "" {
		slog.Warn("event.destroy: invalid id")
		return c.NoContent(400)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("event.destroy: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	err = db.Qry.DeleteEvent(c.Request().Context(), db.DeleteEventParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.destroy: failed to delete event", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "events.notifications.delete_failed"))
		return c.NoContent(500)
	}

	slog.Debug("event.destroy", "id", id)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "events.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID)
		if err != nil {
			slog.Warn("event.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(200)
	}

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	data, err := e.GetIndexData(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.destroy: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, EventIndex(data))
	if err != nil {
		slog.Error("event.destroy: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(200)
}

func (e *Events) CreateParticipant(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if id == "" {
		slog.Warn("participant.create.table: invalid event id")
		return c.NoContent(400)
	}

	var signals participantTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("participant.create.table: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	// Set default expense to 0 if not provided
	expense := signals.FormData.Expense

	_, err = db.Qry.AddParticipant(c.Request().Context(), db.AddParticipantParams{
		GroupID:  groupID,
		EventID:  id,
		MemberID: signals.FormData.MemberID,
		Amount:   signals.FormData.Amount,
		Expense:  expense,
	})
	if err != nil {
		slog.Error("participant.create.table: failed to add participant", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.add_failed"))
		return c.NoContent(500)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.added"))

	data, err := e.GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("participant.create.table: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, EventShow(data))
	if err != nil {
		slog.Error("participant.create.table: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.create.table", "event_id", id, "member_id", signals.FormData.MemberID)
	return c.NoContent(200)
}

func (e *Events) UpdateParticipant(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if eventID == "" {
		slog.Warn("participant.update: invalid event id")
		return c.NoContent(400)
	}

	memberID := c.Param("memberId")
	if memberID == "" {
		slog.Warn("participant.update: invalid member id")
		return c.NoContent(400)
	}

	var signals participantTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Warn("participant.update: failed to read signals", "err", err)
		return c.NoContent(400)
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(422)
	}

	// Set default expense to 0 if not provided
	expense := signals.FormData.Expense

	err = db.Qry.UpdateParticipant(c.Request().Context(), db.UpdateParticipantParams{
		Amount:   signals.FormData.Amount,
		Expense:  expense,
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("participant.update: failed to update participant", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(500)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))

	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)
	data, err := e.GetShowData(c.Request().Context(), groupID, eventID)
	if err != nil {
		slog.Error("participant.update: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, EventShow(data))
	if err != nil {
		slog.Error("participant.update: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)

	slog.Debug("participant.update", "event_id", eventID, "member_id", memberID)
	return c.NoContent(200)
}

func (e *Events) DeleteParticipantTable(c echo.Context) error {
	groupID := getGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if eventID == "" {
		slog.Warn("participant.delete: invalid event id")
		return c.NoContent(400)
	}

	memberID := c.Param("memberId")
	if memberID == "" {
		slog.Warn("participant.delete: invalid member id")
		return c.NoContent(400)
	}

	err := db.Qry.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("participant.delete: failed to remove participant", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.delete_failed"))
		return c.NoContent(500)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.deleted"))

	data, err := e.GetShowData(c.Request().Context(), groupID, eventID)
	if err != nil {
		slog.Error("participant.delete: failed to get data", "err", err)
		return c.NoContent(500)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail
	html, err := utils.RenderComponentStringFor(c, EventShow(data))
	if err != nil {
		slog.Error("participant.delete: failed to render", "err", err)
		return c.NoContent(500)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.delete", "event_id", eventID, "member_id", memberID)
	return c.NoContent(200)
}
