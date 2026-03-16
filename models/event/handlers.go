package event

import (
	"database/sql"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type eventInlineParams struct {
	FormData      eventData        `json:"formData"`
	EventFormData eventData        `json:"eventFormData"`
	TableQuery    utils.TableQuery `json:"tableQuery"`
	Mode          string           `json:"mode"`
}

type modeParams struct {
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
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
	FormData   participantData  `json:"formData"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type participantBulkRowData struct {
	MemberID   string `json:"memberId"`
	MemberName string `json:"memberName"`
	Included   bool   `json:"included"`
	Amount     int64  `json:"amount" validate:"gte=0"`
	Expense    int64  `json:"expense" validate:"gte=0"`
}

type participantBulkParams struct {
	EventFormData     eventData                `json:"eventFormData"`
	WizardEventAmount int64                    `json:"wizardEventAmount"`
	WizardRows        []participantBulkRowData `json:"wizardRows"`
	WizardAmounts     map[string]int64         `json:"wizardAmounts"`
	WizardExpenses    map[string]int64         `json:"wizardExpenses"`
	TableQuery        utils.TableQuery         `json:"tableQuery"`
}

type participantDraftParams struct {
	EventFormData     eventData                `json:"eventFormData"`
	WizardEventAmount int64                    `json:"wizardEventAmount"`
	WizardRows        []participantBulkRowData `json:"wizardRows"`
	WizardAmounts     map[string]int64         `json:"wizardAmounts"`
	WizardExpenses    map[string]int64         `json:"wizardExpenses"`
	TableQuery        utils.TableQuery         `json:"tableQuery"`
}

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
}

func parseParticipantTableQuery(c echo.Context, e *Events) utils.TableQuery {
	return utils.ParseTableQuery(c, staticTableQueryable{spec: e.ParticipantTableQuerySpec()})
}

// Default signal states for resetting forms on success
var (
	defaultEventSignals = map[string]any{
		"mode":      "table",
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

func applyEventIndexTableByRole(data *EventsData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.EventsTable.ActionsWidthRem = 0
	}
}

func applyEventShowTableByRole(data *EventData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.ParticipantsTable.ActionsWidthRem = 0
	}
}

func mergeWizardRows(base []ParticipantWizardRow, allMembers []db.Member, incoming []participantBulkRowData, wizardAmounts map[string]int64, wizardExpenses map[string]int64) []ParticipantWizardRow {
	if len(incoming) == 0 {
		for i := range base {
			if wizardAmounts != nil {
				if amount, ok := wizardAmounts[base[i].MemberID]; ok {
					base[i].Amount = amount
				}
			}
			if wizardExpenses != nil {
				if expense, ok := wizardExpenses[base[i].MemberID]; ok {
					base[i].Expense = expense
				}
			}
		}
		return base
	}

	memberNameByID := make(map[string]string, len(allMembers))
	for _, member := range allMembers {
		memberNameByID[member.ID] = member.Name
	}

	merged := make([]ParticipantWizardRow, 0, len(incoming))
	for _, incomingRow := range incoming {
		memberID := strings.TrimSpace(incomingRow.MemberID)
		if memberID == "" {
			continue
		}

		memberName, ok := memberNameByID[memberID]
		if !ok {
			continue
		}

		merged = append(merged, ParticipantWizardRow{
			MemberID:   memberID,
			MemberName: memberName,
			Included:   incomingRow.Included,
			Amount: func() int64 {
				if wizardAmounts != nil {
					if v, ok := wizardAmounts[memberID]; ok {
						return v
					}
				}
				return incomingRow.Amount
			}(),
			Expense: func() int64 {
				if wizardExpenses != nil {
					if v, ok := wizardExpenses[memberID]; ok {
						return v
					}
				}
				return incomingRow.Expense
			}(),
		})
	}

	if len(merged) == 0 {
		return base
	}

	return merged
}

func buildWizardAddableMembers(allMembers []db.Member, wizardRows []ParticipantWizardRow) []db.Member {
	inDraft := make(map[string]struct{}, len(wizardRows))
	for _, row := range wizardRows {
		inDraft[row.MemberID] = struct{}{}
	}

	addable := make([]db.Member, 0, len(allMembers))
	for _, member := range allMembers {
		if _, exists := inDraft[member.ID]; exists {
			continue
		}
		addable = append(addable, member)
	}

	sort.Slice(addable, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(addable[i].Name))
		right := strings.ToLower(strings.TrimSpace(addable[j].Name))
		if left == right {
			return addable[i].ID < addable[j].ID
		}
		return left < right
	})

	return addable
}

func (e *Events) patchEventShow(c echo.Context, groupID, eventID, userEmail string, query utils.TableQuery, editorMode string, eventForm eventData, wizardEventAmount int64, wizardRows []participantBulkRowData, wizardAmounts map[string]int64, wizardExpenses map[string]int64, wizardError string) error {
	data, err := e.GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		return err
	}

	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	if editorMode != "" {
		data.EditorMode = editorMode
	}

	if eventForm.Title != "" || eventForm.Time != "" || eventForm.Amount > 0 {
		data.Event.Title = eventForm.Title
		data.Event.Time = eventForm.Time
		data.Event.Description = eventForm.Description
		data.Event.Amount = eventForm.Amount
	}

	if wizardEventAmount > 0 {
		data.WizardEventAmount = wizardEventAmount
	}

	if len(wizardRows) > 0 {
		data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, wizardRows, wizardAmounts, wizardExpenses)
	} else if wizardAmounts != nil || wizardExpenses != nil {
		data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, nil, wizardAmounts, wizardExpenses)
	}

	data.WizardAddableMembers = buildWizardAddableMembers(data.AllMembers, data.WizardRows)

	data.WizardError = wizardError

	html, err := utils.RenderHTMLForRequest(c, EventShow(data))
	if err != nil {
		return err
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, eventShowSignals(data, utils.CSRFTokenFromContext(c.Request().Context())))
	return nil
}

func (e *Events) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)
	query := utils.ParseTableQuery(c, e)

	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	slog.Debug("event.index", "event_count", len(data.Events))
	return utils.RenderPage(c, EventIndex(data))
}

func (e *Events) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)
	query := parseParticipantTableQuery(c, e)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	return utils.RenderPage(c, EventShow(data))
}

func (e *Events) Create(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Time = strings.TrimSpace(signals.FormData.Time)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)

	slog.Debug("event.create.table: signals received", "formData", signals.FormData)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		slog.Debug("event.create.table: validation failed", "errors", errs)
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
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
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.create.table", "id", event.ID, "title", event.Title)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "events.notifications.created"))

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	query := utils.NormalizeTableQuery(utils.TableQuery{}, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.create.table: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	html, err := utils.RenderHTMLForRequest(c, EventIndex(data))
	if err != nil {
		slog.Error("event.create.table: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Events) Update(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.update: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.update: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Time = strings.TrimSpace(signals.FormData.Time)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.EventFormData.Title = strings.TrimSpace(signals.EventFormData.Title)
	signals.EventFormData.Time = strings.TrimSpace(signals.EventFormData.Time)
	signals.EventFormData.Description = strings.TrimSpace(signals.EventFormData.Description)

	eventForm := signals.FormData
	if signals.EventFormData.Title != "" || signals.EventFormData.Time != "" || signals.EventFormData.Amount != 0 {
		eventForm = signals.EventFormData
	}

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), eventForm); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
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
		return c.NoContent(http.StatusInternalServerError)
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
		query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
		data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
		if err != nil {
			slog.Error("event.update: failed to get data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		applyEventShowTableByRole(&data, middleware.IsAdmin(c))
		data.UserEmail = userEmail
		html, err := utils.RenderHTMLForRequest(c, EventShow(data))
		if err != nil {
			slog.Error("event.update: failed to render", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.update: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, EventIndex(data))
	if err != nil {
		slog.Error("event.update: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Events) Destroy(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.destroy: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.destroy: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	err = db.Qry.DeleteEvent(c.Request().Context(), db.DeleteEventParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.destroy: failed to delete event", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "events.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.destroy", "id", id)
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "events.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events")
		if err != nil {
			slog.Warn("event.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultEventSignals)
	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, EventIndex(data))
	if err != nil {
		slog.Error("event.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Events) CreateParticipant(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("participant.create.table: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("participant.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.MemberID = strings.TrimSpace(signals.FormData.MemberID)
	signals.FormData.MemberName = strings.TrimSpace(signals.FormData.MemberName)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
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
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.added"))

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("participant.create.table: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, EventShow(data))
	if err != nil {
		slog.Error("participant.create.table: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.create.table", "event_id", id, "member_id", signals.FormData.MemberID)
	return c.NoContent(http.StatusOK)
}

func (e *Events) UpdateParticipant(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.update: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	memberID := c.Param("memberId")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("participant.update: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("participant.update: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.MemberID = strings.TrimSpace(signals.FormData.MemberID)
	signals.FormData.MemberName = strings.TrimSpace(signals.FormData.MemberName)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(participantErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
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
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))

	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)
	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		slog.Error("participant.update: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, EventShow(data))
	if err != nil {
		slog.Error("participant.update: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	slog.Debug("participant.update", "event_id", eventID, "member_id", memberID)
	return c.NoContent(http.StatusOK)
}

func (e *Events) DeleteParticipantTable(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.delete: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	memberID := c.Param("memberId")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("participant.delete: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	err := db.Qry.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("participant.delete: failed to remove participant", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.deleted"))

	query := parseParticipantTableQuery(c, e)
	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		query = utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	}

	data, err := e.GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		slog.Error("participant.delete: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, EventShow(data))
	if err != nil {
		slog.Error("participant.delete: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, defaultParticipantSignals)

	slog.Debug("participant.delete", "event_id", eventID, "member_id", memberID)
	return c.NoContent(http.StatusOK)
}

func (e *Events) OpenParticipantsDraft(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.open: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	query := parseParticipantTableQuery(c, e)
	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		query = utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	}

	if err := e.patchEventShow(c, groupID, eventID, userEmail, query, "edit", eventData{}, 0, nil, nil, nil, ""); err != nil {
		slog.Error("participant.draft.open: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) CancelParticipantsDraft(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.cancel: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	query := parseParticipantTableQuery(c, e)
	var signals participantDraftParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		query = utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	}

	if err := e.patchEventShow(c, groupID, eventID, userEmail, query, "read", eventData{}, 0, nil, nil, nil, ""); err != nil {
		slog.Error("participant.draft.cancel: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) IncludeParticipantsDraftMember(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.include: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	memberID := c.Param("memberId")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("participant.draft.include: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantDraftParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Info("participant.draft.include: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	found := false
	for i := range signals.WizardRows {
		signals.WizardRows[i].MemberID = strings.TrimSpace(signals.WizardRows[i].MemberID)
		if signals.WizardRows[i].MemberID == memberID {
			signals.WizardRows[i].Included = true
			found = true
			break
		}
	}
	if !found {
		amount := int64(0)
		expense := int64(0)
		participants, err := db.Qry.ListParticipantsByEvent(c.Request().Context(), db.ListParticipantsByEventParams{EventID: eventID, GroupID: groupID})
		if err != nil {
			slog.Error("participant.draft.include: failed to list participants", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		for _, participant := range participants {
			if participant.ID != memberID {
				continue
			}
			amount = participant.ParticipantAmount
			expense = participant.ParticipantExpense
			break
		}
		signals.WizardRows = append(signals.WizardRows, participantBulkRowData{MemberID: memberID, Included: true, Amount: amount, Expense: expense})
	}

	if err := e.patchEventShow(c, groupID, eventID, userEmail, query, "edit", signals.EventFormData, signals.WizardEventAmount, signals.WizardRows, signals.WizardAmounts, signals.WizardExpenses, ""); err != nil {
		slog.Error("participant.draft.include: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) ExcludeParticipantsDraftMember(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.exclude: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	memberID := c.Param("memberId")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("participant.draft.exclude: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantDraftParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Info("participant.draft.exclude: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	updatedRows := make([]participantBulkRowData, 0, len(signals.WizardRows))
	for i := range signals.WizardRows {
		signals.WizardRows[i].MemberID = strings.TrimSpace(signals.WizardRows[i].MemberID)
		if signals.WizardRows[i].MemberID == memberID {
			continue
		}
		updatedRows = append(updatedRows, signals.WizardRows[i])
	}
	signals.WizardRows = updatedRows

	if err := e.patchEventShow(c, groupID, eventID, userEmail, query, "edit", signals.EventFormData, signals.WizardEventAmount, signals.WizardRows, signals.WizardAmounts, signals.WizardExpenses, ""); err != nil {
		slog.Error("participant.draft.exclude: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) SaveParticipantsBulk(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.bulk: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantBulkParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Info("participant.bulk: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}

	signals.EventFormData.Title = strings.TrimSpace(signals.EventFormData.Title)
	signals.EventFormData.Time = strings.TrimSpace(signals.EventFormData.Time)
	signals.EventFormData.Description = strings.TrimSpace(signals.EventFormData.Description)

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.EventFormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs), "wizardError": ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error")})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	members, err := db.Qry.ListMembers(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("participant.bulk: failed to list members", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	memberIDs := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberIDs[member.ID] = struct{}{}
	}

	for i := range signals.WizardRows {
		signals.WizardRows[i].MemberID = strings.TrimSpace(signals.WizardRows[i].MemberID)
		signals.WizardRows[i].MemberName = strings.TrimSpace(signals.WizardRows[i].MemberName)
		if signals.WizardAmounts != nil {
			if value, ok := signals.WizardAmounts[signals.WizardRows[i].MemberID]; ok {
				signals.WizardRows[i].Amount = value
			}
		}
		if signals.WizardExpenses != nil {
			if value, ok := signals.WizardExpenses[signals.WizardRows[i].MemberID]; ok {
				signals.WizardRows[i].Expense = value
			}
		}
		if _, ok := memberIDs[signals.WizardRows[i].MemberID]; !ok {
			utils.SSEHub.PatchSignals(c, map[string]any{"wizardError": ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error")})
			return c.NoContent(http.StatusUnprocessableEntity)
		}
		if errs := utils.ValidateWithLocale(c.Request().Context(), signals.WizardRows[i]); errs != nil {
			utils.SSEHub.PatchSignals(c, map[string]any{"wizardError": ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error")})
			return c.NoContent(http.StatusUnprocessableEntity)
		}
		if signals.WizardAmounts != nil {
			if value, ok := signals.WizardAmounts[signals.WizardRows[i].MemberID]; ok && value < 0 {
				utils.SSEHub.PatchSignals(c, map[string]any{"wizardError": ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error")})
				return c.NoContent(http.StatusUnprocessableEntity)
			}
		}
		if signals.WizardExpenses != nil {
			if value, ok := signals.WizardExpenses[signals.WizardRows[i].MemberID]; ok && value < 0 {
				utils.SSEHub.PatchSignals(c, map[string]any{"wizardError": ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error")})
				return c.NoContent(http.StatusUnprocessableEntity)
			}
		}
	}

	tx, err := db.DB.BeginTx(c.Request().Context(), &sql.TxOptions{})
	if err != nil {
		slog.Error("participant.bulk: failed to begin tx", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()
	qtx := db.New(tx)

	_, err = qtx.UpdateEvent(c.Request().Context(), db.UpdateEventParams{
		Title:       signals.EventFormData.Title,
		Time:        signals.EventFormData.Time,
		Description: signals.EventFormData.Description,
		Amount:      signals.EventFormData.Amount,
		ID:          eventID,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("participant.bulk: failed to update event", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	currentParticipants, err := qtx.ListParticipantsByEvent(c.Request().Context(), db.ListParticipantsByEventParams{EventID: eventID, GroupID: groupID})
	if err != nil {
		slog.Error("participant.bulk: failed to list participants", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	currentSet := make(map[string]struct{}, len(currentParticipants))
	for _, participant := range currentParticipants {
		currentSet[participant.ID] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(signals.WizardRows))
	for _, row := range signals.WizardRows {
		if !row.Included {
			continue
		}

		amount := row.Amount
		if signals.WizardAmounts != nil {
			if value, ok := signals.WizardAmounts[row.MemberID]; ok {
				amount = value
			}
		}
		expense := row.Expense
		if signals.WizardExpenses != nil {
			if value, ok := signals.WizardExpenses[row.MemberID]; ok {
				expense = value
			}
		}

		desiredSet[row.MemberID] = struct{}{}
		if _, exists := currentSet[row.MemberID]; exists {
			err = qtx.UpdateParticipant(c.Request().Context(), db.UpdateParticipantParams{
				Amount:   amount,
				Expense:  expense,
				EventID:  eventID,
				MemberID: row.MemberID,
				GroupID:  groupID,
			})
		} else {
			_, err = qtx.AddParticipant(c.Request().Context(), db.AddParticipantParams{
				GroupID:  groupID,
				EventID:  eventID,
				MemberID: row.MemberID,
				Amount:   amount,
				Expense:  expense,
			})
		}
		if err != nil {
			slog.Error("participant.bulk: failed to upsert participant", "err", err, "event_id", eventID, "member_id", row.MemberID)
			utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	for memberID := range currentSet {
		if _, keep := desiredSet[memberID]; keep {
			continue
		}
		err = qtx.RemoveParticipant(c.Request().Context(), db.RemoveParticipantParams{
			EventID:  eventID,
			MemberID: memberID,
			GroupID:  groupID,
		})
		if err != nil {
			slog.Error("participant.bulk: failed to remove participant", "err", err, "event_id", eventID, "member_id", memberID)
			utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	if err = tx.Commit(); err != nil {
		slog.Error("participant.bulk: failed to commit tx", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	if err := e.patchEventShow(c, groupID, eventID, userEmail, query, "read", eventData{}, 0, nil, nil, nil, ""); err != nil {
		slog.Error("participant.bulk: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("participant.bulk", "event_id", eventID, "rows", len(signals.WizardRows))
	return c.NoContent(http.StatusOK)
}
