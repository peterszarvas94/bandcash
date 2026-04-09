package member

import (
	"log/slog"
	"net/http"
	"strings"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
	eventstore "bandcash/models/event/store"
	memberstore "bandcash/models/member/store"
)

// Default signal state for resetting member forms on success
var (
	defaultMemberSignals = map[string]any{
		"mode":      "table",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"name": "", "description": ""},
		"errors":    map[string]any{"name": "", "description": ""},
	}
	memberErrorFields = []string{"name", "description"}
)

func Create(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	var signals memberTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("member.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = strings.TrimSpace(signals.FormData.Name)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	member, err := memberstore.CreateMember(c.Request().Context(), memberstore.CreateMemberParams{
		ID:          utils.GenerateID(utils.PrefixMember),
		GroupID:     groupID,
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
	})
	if err != nil {
		slog.Error("member.create.table: failed to create member", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.create.table", "id", member.ID, "name", member.Name)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.created"))

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/members")
	if err != nil {
		slog.Warn("member.create: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func Update(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Info("member.update: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals memberTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("member.update: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = strings.TrimSpace(signals.FormData.Name)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)

	// Validate
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(memberErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err = memberstore.UpdateMember(c.Request().Context(), memberstore.UpdateMemberParams{
		Name:        signals.FormData.Name,
		Description: signals.FormData.Description,
		ID:          id,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("member.update: failed to update member", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.update", "id", id)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.updated"))

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/members/"+id)
	if err != nil {
		slog.Warn("member.update: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func Destroy(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Info("member.destroy: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("member.destroy: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	err = memberstore.DeleteMember(c.Request().Context(), memberstore.DeleteMemberParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("member.destroy: failed to delete member", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("member.destroy", "id", id)
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "members.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/members")
		if err != nil {
			slog.Warn("member.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultMemberSignals)
	query := utils.NormalizeTableQuery(signals.TableQuery, TableQuerySpec())
	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("member.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = utils.IsAdmin(c)
	data.Signals = memberIndexSignals(utils.TableQuerySignals(data.Query))
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, MemberIndex(data))
	if err != nil {
		slog.Error("member.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func ToggleParticipantPaid(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	memberID := c.Param("id")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("member.toggleParticipantPaid: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	eventID := c.Param("eventId")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("member.toggleParticipantPaid: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("member.toggleParticipantPaid: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	_, err = eventstore.ToggleParticipantPaid(c.Request().Context(), eventstore.ToggleParticipantPaidParams{
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("member.toggleParticipantPaid: failed to toggle paid status", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, MemberEventsTableQuerySpec())
	data, err := GetShowData(c.Request().Context(), groupID, memberID, query)
	if err != nil {
		slog.Error("member.toggleParticipantPaid: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyMemberShowTableByRole(&data, utils.IsAdmin(c))
	data.Signals = memberShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	html, err := utils.RenderHTMLForRequest(c, MemberShow(data))
	if err != nil {
		slog.Error("member.toggleParticipantPaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func OpenParticipantPaidAtDialog(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	memberID := c.Param("id")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("member.openParticipantPaidAtDialog: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	eventID := c.Param("eventId")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("member.openParticipantPaidAtDialog: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.ParseTableQuery(c, staticTableQueryable{spec: MemberEventsTableQuerySpec()})
	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		if utils.SetTabID(c, signals.TabID) {
			query = utils.NormalizeTableQuery(signals.TableQuery, MemberEventsTableQuerySpec())
		}
	}

	data, err := GetShowData(c.Request().Context(), groupID, memberID, query)
	if err != nil {
		slog.Error("member.openParticipantPaidAtDialog: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyMemberShowTableByRole(&data, utils.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	eventTitle := ""
	paidAtValue := ""
	found := false
	for _, event := range data.Events {
		if event.ID != eventID {
			continue
		}
		found = true
		eventTitle = strings.TrimSpace(event.Title)
		if event.ParticipantPaidAt.Valid {
			paidAtValue = utils.FormatDateInput(event.ParticipantPaidAt.String)
		}
		break
	}
	if !found {
		slog.Info("member.openParticipantPaidAtDialog: event not found")
		return c.NoContent(http.StatusBadRequest)
	}

	data.PaidAtDialog = ParticipantPaidAtDialogState{
		Open:        true,
		Title:       ctxi18n.T(c.Request().Context(), "fields.paid_at"),
		Message:     eventTitle,
		EventID:     eventID,
		Value:       paidAtValue,
		SubmitLabel: ctxi18n.T(c.Request().Context(), "table.apply"),
		CancelLabel: ctxi18n.T(c.Request().Context(), "actions.cancel"),
		URL:         "/groups/" + groupID + "/members/" + memberID + "/events/" + eventID + "/paid_at",
		TriggerID:   "member-event-paid-at-edit",
	}
	data.Signals = memberShowSignals(data)

	html, err := utils.RenderHTMLForRequest(c, MemberShow(data))
	if err != nil {
		slog.Error("member.openParticipantPaidAtDialog: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, data.Signals)
	return c.NoContent(http.StatusOK)
}

func UpdateParticipantPaidAt(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	memberID := c.Param("id")
	if !utils.IsValidID(memberID, utils.PrefixMember) {
		slog.Info("member.updateParticipantPaidAt: invalid member id")
		return c.NoContent(http.StatusBadRequest)
	}

	eventID := c.Param("eventId")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		slog.Info("member.updateParticipantPaidAt: invalid event id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals participantPaidAtParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("member.updateParticipantPaidAt: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	_, err = eventstore.UpdateParticipantPaidAt(c.Request().Context(), eventstore.UpdateParticipantPaidAtParams{
		PaidAt:   normalizePaidAtInput(signals.ParticipantPaidAtDialog.Value),
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("member.updateParticipantPaidAt: failed to update paid_at", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(signals.TableQuery, MemberEventsTableQuerySpec())
	data, err := GetShowData(c.Request().Context(), groupID, memberID, query)
	if err != nil {
		slog.Error("member.updateParticipantPaidAt: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyMemberShowTableByRole(&data, utils.IsAdmin(c))
	data.Signals = memberShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	html, err := utils.RenderHTMLForRequest(c, MemberShow(data))
	if err != nil {
		slog.Error("member.updateParticipantPaidAt: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, map[string]any{
		"participantPaidAtDialog": map[string]any{
			"open":      false,
			"fetching":  false,
			"triggerID": "",
		},
	})
	return c.NoContent(http.StatusOK)
}
