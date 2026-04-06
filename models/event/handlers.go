package event

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type eventInlineParams struct {
	TabID                 string           `json:"tab_id"`
	FormData              eventData        `json:"formData"`
	EventFormData         eventData        `json:"eventFormData"`
	TableQuery            utils.TableQuery `json:"tableQuery"`
	Mode                  string           `json:"mode"`
	ParticipantNoteDialog struct {
		MemberID string `json:"memberId"`
		Value    string `json:"value"`
	} `json:"participantNoteDialog"`
}

type modeParams struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type paidAtParams struct {
	TabID        string           `json:"tab_id"`
	Mode         string           `json:"mode"`
	TableQuery   utils.TableQuery `json:"tableQuery"`
	PaidAtDialog struct {
		Value string `json:"value"`
	} `json:"paidAtDialog"`
}

type participantNoteDialogParams struct {
	TabID                 string           `json:"tab_id"`
	TableQuery            utils.TableQuery `json:"tableQuery"`
	ParticipantNoteDialog struct {
		MemberID string `json:"memberId"`
	} `json:"participantNoteDialog"`
}

type eventData struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Time        string `json:"time" validate:"required"`
	Place       string `json:"place" validate:"max=255"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Paid        bool   `json:"paid"`
	PaidAt      string `json:"paidAt"`
}

type participantBulkRowData struct {
	RowID      string `json:"rowId"`
	MemberID   string `json:"memberId"`
	MemberName string `json:"memberName"`
	Included   bool   `json:"included"`
	Amount     int64  `json:"amount" validate:"gte=0"`
	Expense    int64  `json:"expense" validate:"gte=0"`
	Note       string `json:"note" validate:"max=1000"`
	Paid       bool   `json:"paid"`
	PaidAt     string `json:"paidAt"`
}

type participantWizardSignals struct {
	EventAmount int64                    `json:"eventAmount"`
	Rows        []participantBulkRowData `json:"rows"`
	MemberIDs   map[string]string        `json:"memberIds"`
	Amounts     map[string]int64         `json:"amounts"`
	Expenses    map[string]int64         `json:"expenses"`
	Notes       map[string]string        `json:"notes"`
	Paids       map[string]bool          `json:"paids"`
	PaidAts     map[string]string        `json:"paidAts"`
	Total       int64                    `json:"total"`
	Leftover    int64                    `json:"leftover"`
	Error       string                   `json:"error"`
}

type participantBulkParams struct {
	TabID         string                   `json:"tab_id"`
	EventFormData eventData                `json:"eventFormData"`
	Wizard        participantWizardSignals `json:"wizard"`
	TableQuery    utils.TableQuery         `json:"tableQuery"`
}

type participantDraftParams struct {
	TabID         string                   `json:"tab_id"`
	EventFormData eventData                `json:"eventFormData"`
	Wizard        participantWizardSignals `json:"wizard"`
	TableQuery    utils.TableQuery         `json:"tableQuery"`
}

type participantDraftRowParams struct {
	TabID           string                   `json:"tab_id"`
	DraftRowsAction string                   `json:"draftRowsAction"`
	DraftRowsRowID  string                   `json:"draftRowsRowId"`
	EventFormData   eventData                `json:"eventFormData"`
	Wizard          participantWizardSignals `json:"wizard"`
	TableQuery      utils.TableQuery         `json:"tableQuery"`
}

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
}

func parseParticipantTableQuery(c echo.Context, e *Events) utils.TableQuery {
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: e.ParticipantTableQuerySpec()})
	query.Page = 1
	query.Search = ""
	query.PageSize = utils.DefaultTablePageSize
	return query
}

func normalizePaidAtInput(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	formatted := utils.FormatDateInput(trimmed)
	if formatted != "" {
		return formatted
	}

	return trimmed
}

func paidAtArg(isPaid bool, paidAt string) sql.NullString {
	if !isPaid {
		return sql.NullString{}
	}

	normalized := normalizePaidAtInput(paidAt)
	if normalized == "" {
		return sql.NullString{String: "", Valid: true}
	}

	return sql.NullString{String: normalized, Valid: true}
}

// Default signal states for resetting forms on success
var (
	defaultEventSignals = map[string]any{
		"mode":      "table",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": 0, "paid": false, "paidAt": ""},
		"errors":    map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": ""},
	}
	// Error field lists for validation
	eventErrorFields = []string{"title", "time", "place", "description", "amount"}
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

func mergeWizardRows(base []ParticipantWizardRow, allMembers []db.Member, incoming []participantBulkRowData, wizardMemberIDs map[string]string, wizardAmounts map[string]int64, wizardExpenses map[string]int64, wizardNotes map[string]string, wizardPaids map[string]bool, wizardPaidAts map[string]string) []ParticipantWizardRow {
	if len(incoming) == 0 {
		for i := range base {
			if wizardMemberIDs != nil {
				if memberID, ok := wizardMemberIDs[base[i].RowID]; ok {
					base[i].MemberID = strings.TrimSpace(memberID)
				}
			}
			if wizardAmounts != nil {
				if amount, ok := wizardAmounts[base[i].RowID]; ok {
					base[i].Amount = amount
				}
			}
			if wizardExpenses != nil {
				if expense, ok := wizardExpenses[base[i].RowID]; ok {
					base[i].Expense = expense
				}
			}
			if wizardNotes != nil {
				if note, ok := wizardNotes[base[i].RowID]; ok {
					base[i].Note = strings.TrimSpace(note)
				}
			}
			if wizardPaids != nil {
				if paid, ok := wizardPaids[base[i].RowID]; ok {
					base[i].Paid = paid
				}
			}
			if wizardPaidAts != nil {
				if paidAt, ok := wizardPaidAts[base[i].RowID]; ok {
					base[i].PaidAt = normalizePaidAtInput(paidAt)
				}
			}
		}

		memberNameByID := make(map[string]string, len(allMembers))
		for _, member := range allMembers {
			memberNameByID[member.ID] = member.Name
		}
		for i := range base {
			if base[i].MemberID == "" {
				base[i].MemberName = ""
				continue
			}
			base[i].MemberName = memberNameByID[base[i].MemberID]
		}

		return base
	}

	memberNameByID := make(map[string]string, len(allMembers))
	for _, member := range allMembers {
		memberNameByID[member.ID] = member.Name
	}

	merged := make([]ParticipantWizardRow, 0, len(incoming))
	for _, incomingRow := range incoming {
		rowID := strings.TrimSpace(incomingRow.RowID)
		if rowID == "" {
			rowID = utils.GenerateID(utils.PrefixParticipant)
		}

		memberID := strings.TrimSpace(incomingRow.MemberID)
		if wizardMemberIDs != nil {
			if value, ok := wizardMemberIDs[rowID]; ok {
				memberID = strings.TrimSpace(value)
			}
		}

		amount := incomingRow.Amount
		if wizardAmounts != nil {
			if value, ok := wizardAmounts[rowID]; ok {
				amount = value
			}
		}

		expense := incomingRow.Expense
		if wizardExpenses != nil {
			if value, ok := wizardExpenses[rowID]; ok {
				expense = value
			}
		}

		note := strings.TrimSpace(incomingRow.Note)
		if wizardNotes != nil {
			if value, ok := wizardNotes[rowID]; ok {
				note = strings.TrimSpace(value)
			}
		}

		paid := incomingRow.Paid
		if wizardPaids != nil {
			if value, ok := wizardPaids[rowID]; ok {
				paid = value
			}
		}

		paidAt := normalizePaidAtInput(incomingRow.PaidAt)
		if wizardPaidAts != nil {
			if value, ok := wizardPaidAts[rowID]; ok {
				paidAt = normalizePaidAtInput(value)
			}
		}

		memberName := ""
		if memberID == "" {
		} else {
			memberName = memberNameByID[memberID]
		}

		merged = append(merged, ParticipantWizardRow{
			RowID:      rowID,
			MemberID:   memberID,
			MemberName: memberName,
			Included:   incomingRow.Included,
			Amount:     amount,
			Expense:    expense,
			Note:       note,
			Paid:       paid,
			PaidAt:     paidAt,
		})
	}

	return merged
}

func patchWizardError(c echo.Context, wizard participantWizardSignals, message string) {
	utils.SSEHub.PatchSignals(c, map[string]any{
		"wizard": map[string]any{
			"eventAmount": wizard.EventAmount,
			"rows":        wizard.Rows,
			"memberIds":   wizard.MemberIDs,
			"amounts":     wizard.Amounts,
			"expenses":    wizard.Expenses,
			"notes":       wizard.Notes,
			"paids":       wizard.Paids,
			"paidAts":     wizard.PaidAts,
			"total":       wizard.Total,
			"leftover":    wizard.Leftover,
			"error":       message,
		},
		"errors": map[string]any{
			"memberId": message,
		},
	})
}

func (e *Events) patchEventShow(c echo.Context, groupID, eventID string, query utils.TableQuery, editorMode string, eventForm eventData, wizardEventAmount int64, wizardRows []participantBulkRowData, wizardMemberIDs map[string]string, wizardAmounts map[string]int64, wizardExpenses map[string]int64, wizardNotes map[string]string, wizardPaids map[string]bool, wizardPaidAts map[string]string, wizardError string) error {
	data, err := e.GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		return err
	}

	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	if editorMode != "" {
		data.EditorMode = editorMode
	}
	if data.EditorMode == "edit" {
		if len(data.Breadcrumbs) > 0 {
			data.Breadcrumbs[len(data.Breadcrumbs)-1].Href = "/groups/" + groupID + "/events/" + eventID
		}
		data.Breadcrumbs = append(data.Breadcrumbs, utils.Crumb{Label: ctxi18n.T(c.Request().Context(), "events.edit")})
	}

	if eventForm.Title != "" || eventForm.Time != "" || eventForm.Place != "" || eventForm.Amount > 0 {
		data.Event.Title = eventForm.Title
		data.Event.Time = eventForm.Time
		data.Event.Place = eventForm.Place
		data.Event.Description = eventForm.Description
		data.Event.Amount = eventForm.Amount
		if eventForm.Paid {
			data.Event.Paid = 1
		} else {
			data.Event.Paid = 0
		}
		if eventForm.Paid {
			if eventForm.PaidAt != "" {
				data.Event.PaidAt = sql.NullString{String: eventForm.PaidAt, Valid: true}
			}
		} else {
			data.Event.PaidAt = sql.NullString{}
		}
	}

	if wizardEventAmount > 0 {
		data.WizardEventAmount = wizardEventAmount
	}

	if wizardRows != nil {
		if len(wizardRows) == 0 {
			data.WizardRows = []ParticipantWizardRow{}
		} else {
			data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, wizardRows, wizardMemberIDs, wizardAmounts, wizardExpenses, wizardNotes, wizardPaids, wizardPaidAts)
		}
	} else if wizardMemberIDs != nil || wizardAmounts != nil || wizardExpenses != nil || wizardNotes != nil || wizardPaids != nil || wizardPaidAts != nil {
		data.WizardRows = mergeWizardRows(data.WizardRows, data.AllMembers, nil, wizardMemberIDs, wizardAmounts, wizardExpenses, wizardNotes, wizardPaids, wizardPaidAts)
	}

	data.WizardError = wizardError
	data.Signals = eventShowSignals(data)

	var html string
	if data.EditorMode == "edit" {
		html, err = utils.RenderHTMLForRequest(c, EventEditPage(data))
	} else {
		html, err = utils.RenderHTMLForRequest(c, EventShowPage(data))
	}
	if err != nil {
		return err
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, eventShowSignals(data))
	return nil
}

func (e *Events) IndexPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := utils.ParseTableQuery(c, e)

	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	slog.Debug("event.index", "event_count", len(data.Events))
	return utils.RenderPage(c, EventIndexPage(data))
}

func (e *Events) ShowPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
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
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, EventShowPage(data))
}

func (e *Events) NewEventPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.new_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := NewEventPageData{
		Title: ctxi18n.T(c.Request().Context(), "events.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "events.add")},
		},
		GroupID: groupID,
		Signals: map[string]any{
			"formData": map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": 0, "paid": false, "paidAt": ""},
			"errors":   map[string]any{"title": "", "time": "", "place": "", "description": "", "amount": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    middleware.IsSuperadmin(c),
	}
	return utils.RenderPage(c, EventNewPage(data))
}

func (e *Events) EditEventPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := parseParticipantTableQuery(c, e)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.edit_page: failed to get data", "group_id", groupID, "event_id", id, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	if len(data.Breadcrumbs) > 0 {
		data.Breadcrumbs[len(data.Breadcrumbs)-1].Href = "/groups/" + groupID + "/events/" + id
	}
	data.Breadcrumbs = append(data.Breadcrumbs, utils.Crumb{Label: ctxi18n.T(c.Request().Context(), "events.edit")})
	data.EditorMode = "edit"
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, EventEditPage(data))
}

func (e *Events) Create(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Time = strings.TrimSpace(signals.FormData.Time)
	signals.FormData.Place = strings.TrimSpace(signals.FormData.Place)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.PaidAt = normalizePaidAtInput(signals.FormData.PaidAt)

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
		Place:       signals.FormData.Place,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
		Paid: func() int64 {
			if signals.FormData.Paid {
				return 1
			}
			return 0
		}(),
		PaidAt: paidAtArg(signals.FormData.Paid, signals.FormData.PaidAt),
	})
	if err != nil {
		slog.Error("event.create.table: failed to create event", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.create: created", "id", event.ID, "title", event.Title)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.created"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events")
	if err != nil {
		slog.Warn("event.create: failed to redirect", "err", err)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) Update(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

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
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Time = strings.TrimSpace(signals.FormData.Time)
	signals.FormData.Place = strings.TrimSpace(signals.FormData.Place)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.PaidAt = normalizePaidAtInput(signals.FormData.PaidAt)
	signals.EventFormData.Title = strings.TrimSpace(signals.EventFormData.Title)
	signals.EventFormData.Time = strings.TrimSpace(signals.EventFormData.Time)
	signals.EventFormData.Place = strings.TrimSpace(signals.EventFormData.Place)
	signals.EventFormData.Description = strings.TrimSpace(signals.EventFormData.Description)
	signals.EventFormData.PaidAt = normalizePaidAtInput(signals.EventFormData.PaidAt)

	eventForm := signals.FormData
	if signals.EventFormData.Title != "" || signals.EventFormData.Time != "" || signals.EventFormData.Place != "" || signals.EventFormData.Amount != 0 {
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
		Place:       eventForm.Place,
		Description: eventForm.Description,
		Amount:      eventForm.Amount,
		Paid: func() int64 {
			if eventForm.Paid {
				return 1
			}
			return 0
		}(),
		PaidAt:  paidAtArg(eventForm.Paid, eventForm.PaidAt),
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.update: failed to update event", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.update", "id", id)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.updated"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events/"+id)
	if err != nil {
		slog.Warn("event.update: failed to redirect", "err", err)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) Destroy(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

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
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	err = db.Qry.DeleteEvent(c.Request().Context(), db.DeleteEventParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.destroy: failed to delete event", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.destroy", "id", id)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events")
		if err != nil {
			slog.Warn("event.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultEventSignals)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, EventIndexPage(data))
	if err != nil {
		slog.Error("event.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Events) ToggleParticipantPaid(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.togglePaid: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	memberID := c.Param("memberId")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("participant.togglePaid: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("participant.togglePaid: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	_, err = db.Qry.ToggleParticipantPaid(c.Request().Context(), db.ToggleParticipantPaidParams{
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("participant.togglePaid: failed to toggle paid status", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("participant.togglePaid", "event_id", eventID, "member_id", memberID)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, eventID, query)
	if err != nil {
		slog.Error("participant.togglePaid: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
	if err != nil {
		slog.Error("participant.togglePaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func (e *Events) OpenParticipantsDraft(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.open: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events/"+eventID+"/edit"); err != nil {
		slog.Warn("participant.draft.open: failed to redirect", "err", err)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) CancelParticipantsDraft(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.cancel: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantDraftParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events/"+eventID); err != nil {
		slog.Warn("participant.draft.cancel: failed to redirect", "err", err)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) UpdateParticipantsDraftRows(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("participant.draft.rows: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantDraftRowParams
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		slog.Info("participant.draft.rows: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	signals.EventFormData.Title = strings.TrimSpace(signals.EventFormData.Title)
	signals.EventFormData.Time = strings.TrimSpace(signals.EventFormData.Time)
	signals.EventFormData.Place = strings.TrimSpace(signals.EventFormData.Place)
	signals.EventFormData.Description = strings.TrimSpace(signals.EventFormData.Description)
	signals.EventFormData.PaidAt = normalizePaidAtInput(signals.EventFormData.PaidAt)

	rows := append([]participantBulkRowData(nil), signals.Wizard.Rows...)
	for i := range rows {
		rows[i].RowID = strings.TrimSpace(rows[i].RowID)
		if rows[i].RowID == "" {
			rows[i].RowID = strings.TrimSpace(rows[i].MemberID)
		}
	}
	if signals.Wizard.MemberIDs == nil {
		signals.Wizard.MemberIDs = map[string]string{}
	}
	if signals.Wizard.Amounts == nil {
		signals.Wizard.Amounts = map[string]int64{}
	}
	if signals.Wizard.Expenses == nil {
		signals.Wizard.Expenses = map[string]int64{}
	}
	if signals.Wizard.Notes == nil {
		signals.Wizard.Notes = map[string]string{}
	}
	if signals.Wizard.Paids == nil {
		signals.Wizard.Paids = map[string]bool{}
	}
	if signals.Wizard.PaidAts == nil {
		signals.Wizard.PaidAts = map[string]string{}
	}

	switch strings.TrimSpace(signals.DraftRowsAction) {
	case "add":
		rowID := utils.GenerateID(utils.PrefixParticipant)
		rows = append(rows, participantBulkRowData{RowID: rowID, Included: true})
		signals.Wizard.MemberIDs[rowID] = ""
		signals.Wizard.Amounts[rowID] = 0
		signals.Wizard.Expenses[rowID] = 0
		signals.Wizard.Notes[rowID] = ""
		signals.Wizard.Paids[rowID] = false
		signals.Wizard.PaidAts[rowID] = ""
	case "copy":
		targetRowID := strings.TrimSpace(signals.DraftRowsRowID)
		if targetRowID == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		sourceIndex := -1
		for i := range rows {
			if rows[i].RowID == targetRowID {
				sourceIndex = i
				break
			}
		}
		if sourceIndex < 0 {
			return c.NoContent(http.StatusBadRequest)
		}
		source := rows[sourceIndex]
		newRowID := utils.GenerateID(utils.PrefixParticipant)
		sourceAmount := source.Amount
		if value, ok := signals.Wizard.Amounts[source.RowID]; ok {
			sourceAmount = value
		}
		sourceExpense := source.Expense
		if value, ok := signals.Wizard.Expenses[source.RowID]; ok {
			sourceExpense = value
		}
		sourceNote := strings.TrimSpace(source.Note)
		if value, ok := signals.Wizard.Notes[source.RowID]; ok {
			sourceNote = strings.TrimSpace(value)
		}
		sourcePaid := source.Paid
		if value, ok := signals.Wizard.Paids[source.RowID]; ok {
			sourcePaid = value
		}
		sourcePaidAt := normalizePaidAtInput(source.PaidAt)
		if value, ok := signals.Wizard.PaidAts[source.RowID]; ok {
			sourcePaidAt = normalizePaidAtInput(value)
		}
		rows = append(rows, participantBulkRowData{
			RowID:    newRowID,
			MemberID: "",
			Included: true,
			Amount:   sourceAmount,
			Expense:  sourceExpense,
			Note:     sourceNote,
			Paid:     sourcePaid,
			PaidAt:   sourcePaidAt,
		})
		signals.Wizard.MemberIDs[newRowID] = ""
		signals.Wizard.Amounts[newRowID] = sourceAmount
		signals.Wizard.Expenses[newRowID] = sourceExpense
		signals.Wizard.Notes[newRowID] = sourceNote
		signals.Wizard.Paids[newRowID] = sourcePaid
		signals.Wizard.PaidAts[newRowID] = sourcePaidAt
	case "remove":
		targetRowID := strings.TrimSpace(signals.DraftRowsRowID)
		if targetRowID == "" {
			return c.NoContent(http.StatusBadRequest)
		}
		updatedRows := make([]participantBulkRowData, 0, len(rows))
		for _, row := range rows {
			if row.RowID == targetRowID {
				continue
			}
			updatedRows = append(updatedRows, row)
		}
		rows = updatedRows
		delete(signals.Wizard.MemberIDs, targetRowID)
		delete(signals.Wizard.Amounts, targetRowID)
		delete(signals.Wizard.Expenses, targetRowID)
		delete(signals.Wizard.Notes, targetRowID)
		delete(signals.Wizard.Paids, targetRowID)
		delete(signals.Wizard.PaidAts, targetRowID)
	default:
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
	if err := e.patchEventShow(c, groupID, eventID, query, "edit", signals.EventFormData, signals.Wizard.EventAmount, rows, signals.Wizard.MemberIDs, signals.Wizard.Amounts, signals.Wizard.Expenses, signals.Wizard.Notes, signals.Wizard.Paids, signals.Wizard.PaidAts, ""); err != nil {
		slog.Error("participant.draft.rows: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

func (e *Events) SaveParticipantsBulk(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

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
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	signals.EventFormData.Title = strings.TrimSpace(signals.EventFormData.Title)
	signals.EventFormData.Time = strings.TrimSpace(signals.EventFormData.Time)
	signals.EventFormData.Place = strings.TrimSpace(signals.EventFormData.Place)
	signals.EventFormData.Description = strings.TrimSpace(signals.EventFormData.Description)
	signals.EventFormData.PaidAt = normalizePaidAtInput(signals.EventFormData.PaidAt)

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.EventFormData); errs != nil {
		patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(eventErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	members, err := db.Qry.ListMembers(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("participant.bulk: failed to list members", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	memberIDs := make(map[string]struct{}, len(members))
	for _, member := range members {
		memberIDs[member.ID] = struct{}{}
	}

	memberIDsSeen := make(map[string]struct{}, len(signals.Wizard.Rows))
	normalizedRows := make([]participantBulkRowData, 0, len(signals.Wizard.Rows))
	if signals.Wizard.MemberIDs == nil {
		signals.Wizard.MemberIDs = map[string]string{}
	}
	for i := range signals.Wizard.Rows {
		signals.Wizard.Rows[i].RowID = strings.TrimSpace(signals.Wizard.Rows[i].RowID)
		if signals.Wizard.Rows[i].RowID == "" {
			patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		signals.Wizard.Rows[i].MemberID = strings.TrimSpace(signals.Wizard.Rows[i].MemberID)
		if memberID, ok := signals.Wizard.MemberIDs[signals.Wizard.Rows[i].RowID]; ok {
			signals.Wizard.Rows[i].MemberID = strings.TrimSpace(memberID)
		}
		signals.Wizard.Rows[i].MemberName = strings.TrimSpace(signals.Wizard.Rows[i].MemberName)
		signals.Wizard.Rows[i].Note = strings.TrimSpace(signals.Wizard.Rows[i].Note)
		signals.Wizard.Rows[i].PaidAt = normalizePaidAtInput(signals.Wizard.Rows[i].PaidAt)
		if signals.Wizard.Amounts != nil {
			if value, ok := signals.Wizard.Amounts[signals.Wizard.Rows[i].RowID]; ok {
				signals.Wizard.Rows[i].Amount = value
			}
		}
		if signals.Wizard.Expenses != nil {
			if value, ok := signals.Wizard.Expenses[signals.Wizard.Rows[i].RowID]; ok {
				signals.Wizard.Rows[i].Expense = value
			}
		}
		if signals.Wizard.Notes != nil {
			if value, ok := signals.Wizard.Notes[signals.Wizard.Rows[i].RowID]; ok {
				signals.Wizard.Rows[i].Note = strings.TrimSpace(value)
			}
		}
		if signals.Wizard.Paids != nil {
			if value, ok := signals.Wizard.Paids[signals.Wizard.Rows[i].RowID]; ok {
				signals.Wizard.Rows[i].Paid = value
			}
		}
		if signals.Wizard.PaidAts != nil {
			if value, ok := signals.Wizard.PaidAts[signals.Wizard.Rows[i].RowID]; ok {
				signals.Wizard.Rows[i].PaidAt = normalizePaidAtInput(value)
			}
		}

		if signals.Wizard.Rows[i].MemberID == "" {
			if signals.Wizard.Rows[i].Amount == 0 && signals.Wizard.Rows[i].Expense == 0 && signals.Wizard.Rows[i].Note == "" && !signals.Wizard.Rows[i].Paid && signals.Wizard.Rows[i].PaidAt == "" {
				continue
			}
			patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		if _, ok := memberIDs[signals.Wizard.Rows[i].MemberID]; !ok {
			patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		if _, exists := memberIDsSeen[signals.Wizard.Rows[i].MemberID]; exists {
			patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
			return c.NoContent(http.StatusUnprocessableEntity)
		}
		memberIDsSeen[signals.Wizard.Rows[i].MemberID] = struct{}{}

		if errs := utils.ValidateWithLocale(c.Request().Context(), signals.Wizard.Rows[i]); errs != nil {
			patchWizardError(c, signals.Wizard, ctxi18n.T(c.Request().Context(), "participants.bulk_validation_error"))
			return c.NoContent(http.StatusUnprocessableEntity)
		}

		normalizedRows = append(normalizedRows, signals.Wizard.Rows[i])
	}
	signals.Wizard.Rows = normalizedRows

	tx, err := db.DB.BeginTx(c.Request().Context(), &sql.TxOptions{})
	if err != nil {
		slog.Error("participant.bulk: failed to begin tx", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	defer tx.Rollback()
	qtx := db.New(tx)

	_, err = qtx.UpdateEvent(c.Request().Context(), db.UpdateEventParams{
		Title:       signals.EventFormData.Title,
		Time:        signals.EventFormData.Time,
		Place:       signals.EventFormData.Place,
		Description: signals.EventFormData.Description,
		Amount:      signals.EventFormData.Amount,
		Paid: func() int64 {
			if signals.EventFormData.Paid {
				return 1
			}
			return 0
		}(),
		PaidAt:  paidAtArg(signals.EventFormData.Paid, signals.EventFormData.PaidAt),
		ID:      eventID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("participant.bulk: failed to update event", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	currentParticipants, err := qtx.ListParticipantsByEvent(c.Request().Context(), db.ListParticipantsByEventParams{EventID: eventID, GroupID: groupID})
	if err != nil {
		slog.Error("participant.bulk: failed to list participants", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	currentSet := make(map[string]struct{}, len(currentParticipants))
	for _, participant := range currentParticipants {
		currentSet[participant.ID] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(signals.Wizard.Rows))
	for _, row := range signals.Wizard.Rows {
		if !row.Included {
			continue
		}

		amount := row.Amount
		expense := row.Expense
		paid := row.Paid
		paidAt := normalizePaidAtInput(row.PaidAt)

		desiredSet[row.MemberID] = struct{}{}
		if _, exists := currentSet[row.MemberID]; exists {
			err = qtx.UpdateParticipant(c.Request().Context(), db.UpdateParticipantParams{
				Amount:  amount,
				Expense: expense,
				Note:    row.Note,
				Paid: func() int64 {
					if paid {
						return 1
					}
					return 0
				}(),
				PaidAt:   paidAtArg(paid, paidAt),
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
				Note:     row.Note,
				Paid: func() int64 {
					if paid {
						return 1
					}
					return 0
				}(),
				PaidAt: paidAtArg(paid, paidAt),
			})
		}
		if err != nil {
			slog.Error("participant.bulk: failed to upsert participant", "err", err, "event_id", eventID, "member_id", row.MemberID)
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
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
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	if err = tx.Commit(); err != nil {
		slog.Error("participant.bulk: failed to commit tx", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	if err := utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events/"+eventID); err != nil {
		slog.Warn("participant.bulk: failed to redirect", "err", err)
	}

	slog.Debug("participant.bulk", "event_id", eventID, "rows", len(signals.Wizard.Rows))
	return c.NoContent(http.StatusOK)
}

func (e *Events) openPaidAtDialog(c echo.Context, groupID, id string, tableQuery utils.TableQuery) error {
	query := utils.NormalizeTableQuery(tableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.openPaidAtDialog: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	data.PaidAtDialog = PaidAtDialogState{
		Open:  true,
		Title: ctxi18n.T(c.Request().Context(), "fields.paid_at"),
		Value: func() string {
			if data.Event.PaidAt.Valid {
				return utils.FormatDateInput(data.Event.PaidAt.String)
			}
			return ""
		}(),
		SubmitLabel: ctxi18n.T(c.Request().Context(), "table.apply"),
		CancelLabel: ctxi18n.T(c.Request().Context(), "actions.cancel"),
		URL:         "/groups/" + groupID + "/events/" + id + "/paid_at",
		TriggerID:   "event-income-paid-at-edit",
	}
	data.Signals = eventShowSignals(data)

	html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
	if err != nil {
		slog.Error("event.openPaidAtDialog: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, data.Signals)
	return c.NoContent(http.StatusOK)
}

func (e *Events) openParticipantNoteDialog(c echo.Context, groupID, id, memberID string, tableQuery utils.TableQuery) error {
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("event.openParticipantNoteDialog: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.NormalizeTableQuery(tableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.openParticipantNoteDialog: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	participantName := ""
	participantNote := ""
	found := false
	for _, participant := range data.Participants {
		if participant.ID == memberID {
			participantName = participant.Name
			participantNote = strings.TrimSpace(participant.ParticipantNote)
			found = true
			break
		}
	}
	if !found {
		slog.Info("event.openParticipantNoteDialog: participant not found")
		return c.NoContent(http.StatusBadRequest)
	}

	isAdmin := middleware.IsAdmin(c)
	data.ParticipantNoteDialog = ParticipantNoteDialogState{
		Open:        true,
		ReadOnly:    !isAdmin,
		Title:       ctxi18n.T(c.Request().Context(), "participants.note"),
		Message:     participantName,
		MemberID:    memberID,
		Value:       participantNote,
		SubmitLabel: ctxi18n.T(c.Request().Context(), "table.apply"),
		CancelLabel: ctxi18n.T(c.Request().Context(), "actions.cancel"),
		URL:         "/groups/" + groupID + "/events/" + id + "/note",
		TriggerID:   "event-participant-note-dialog",
	}
	data.Signals = eventShowSignals(data)

	html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
	if err != nil {
		slog.Error("event.openParticipantNoteDialog: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, data.Signals)
	return c.NoContent(http.StatusOK)
}

func (e *Events) patchUpdatePaidAt(c echo.Context, groupID, id, _ string, tableQuery utils.TableQuery, value string) error {
	paidAt := normalizePaidAtInput(value)

	_, err := db.Qry.UpdateEventPaidAt(c.Request().Context(), db.UpdateEventPaidAtParams{
		PaidAt:  paidAt,
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.updatePaidAt: failed to update paid_at", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.updated"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(tableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.updatePaidAt: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
	if err != nil {
		slog.Error("event.updatePaidAt: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, map[string]any{
		"paidAtDialog": map[string]any{
			"open":      false,
			"fetching":  false,
			"triggerID": "",
		},
	})
	return c.NoContent(http.StatusOK)
}

func (e *Events) patchUpdateParticipantNote(c echo.Context, groupID, id string, tableQuery utils.TableQuery, memberID, value string) error {
	if !middleware.IsAdmin(c) {
		return c.NoContent(http.StatusForbidden)
	}
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("event.updateParticipantNote: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	err := db.Qry.UpdateParticipantNote(c.Request().Context(), db.UpdateParticipantNoteParams{
		Note:     strings.TrimSpace(value),
		EventID:  id,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("event.updateParticipantNote: failed to update note", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(tableQuery, e.ParticipantTableQuerySpec())
	data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.updateParticipantNote: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
	if err != nil {
		slog.Error("event.updateParticipantNote: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, map[string]any{
		"participantNoteDialog": map[string]any{
			"open":      false,
			"fetching":  false,
			"triggerID": "",
		},
	})
	return c.NoContent(http.StatusOK)
}

func (e *Events) patchTogglePaid(c echo.Context, groupID, id, mode string, tableQuery utils.TableQuery) error {
	_, err := db.Qry.ToggleEventPaid(c.Request().Context(), db.ToggleEventPaidParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("event.togglePaid: failed to toggle paid status", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("event.togglePaid", "id", id)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	if mode == "single" {
		query := utils.NormalizeTableQuery(tableQuery, e.ParticipantTableQuerySpec())
		data, err := e.GetShowData(c.Request().Context(), groupID, id, query)
		if err != nil {
			slog.Error("event.togglePaid: failed to get data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		applyEventShowTableByRole(&data, middleware.IsAdmin(c))
		data.Signals = eventShowSignals(data)
		data.IsAuthenticated = true
		data.IsSuperAdmin = middleware.IsSuperadmin(c)
		html, err := utils.RenderHTMLForRequest(c, EventShowPage(data))
		if err != nil {
			slog.Error("event.togglePaid: failed to render", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	query := utils.NormalizeTableQuery(tableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.togglePaid: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = eventIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, EventIndexPage(data))
	if err != nil {
		slog.Error("event.togglePaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func (e *Events) OpenPaidAtPrompt(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.openPaidAtPrompt: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}
	query := parseParticipantTableQuery(c, e)
	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		if utils.SetTabID(c, signals.TabID) {
			query = utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
		}
	} else if tabID := strings.TrimSpace(c.QueryParam("tab_id")); utils.IsValidID(tabID, "tab") {
		_ = utils.SetTabID(c, tabID)
	}

	return e.openPaidAtDialog(c, groupID, id, query)
}

func (e *Events) OpenParticipantNoteDialog(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.openParticipantNoteDialog: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}
	memberID := strings.TrimSpace(c.QueryParam("member_id"))
	query := parseParticipantTableQuery(c, e)
	var signals participantNoteDialogParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		if utils.SetTabID(c, signals.TabID) {
			query = utils.NormalizeTableQuery(signals.TableQuery, e.ParticipantTableQuerySpec())
		}
		if memberID == "" {
			memberID = strings.TrimSpace(signals.ParticipantNoteDialog.MemberID)
		}
	} else if tabID := strings.TrimSpace(c.QueryParam("tab_id")); utils.IsValidID(tabID, "tab") {
		_ = utils.SetTabID(c, tabID)
	}

	return e.openParticipantNoteDialog(c, groupID, id, memberID, query)
}

func (e *Events) UpdatePaidAt(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.updatePaidAt: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals paidAtParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.updatePaidAt: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	return e.patchUpdatePaidAt(c, groupID, id, signals.Mode, signals.TableQuery, signals.PaidAtDialog.Value)
}

func (e *Events) UpdateParticipantNote(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.updateParticipantNote: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals eventInlineParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.updateParticipantNote: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	return e.patchUpdateParticipantNote(c, groupID, id, signals.TableQuery, signals.ParticipantNoteDialog.MemberID, signals.ParticipantNoteDialog.Value)
}

func (e *Events) TogglePaid(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.togglePaid: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("event.togglePaid: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	return e.patchTogglePaid(c, groupID, id, signals.Mode, signals.TableQuery)
}
