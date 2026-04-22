package group

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	eventstore "bandcash/models/event/data"
	expensestore "bandcash/models/expense/data"
	groupstore "bandcash/models/group/data"
)

type Group struct {
	model *GroupModel
}

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
}

var errAtLeastOneAdmin = errors.New("at least one admin required")

const paymentsFadeDuration = 200 * time.Millisecond

func usersTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "email",
		DefaultDir:   "asc",
		AllowedSorts: []string{"email", "role", "status", "createdAt"},
	})
}

func toReceivePaymentsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "date",
		DefaultDir:   "asc",
		AllowedSorts: []string{"title", "amount", "paid", "paid_at", "date"},
	})
}

func toPayPaymentsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "date",
		DefaultDir:   "asc",
		AllowedSorts: []string{"type", "title", "amount", "paid", "paid_at", "date"},
	})
}

func recentIncomePaymentsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "paid_at",
		DefaultDir:   "desc",
		AllowedSorts: []string{"title", "amount", "paid", "paid_at", "date"},
	})
}

func recentOutgoingPaymentsTableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "paid_at",
		DefaultDir:   "desc",
		AllowedSorts: []string{"type", "title", "amount", "paid", "paid_at", "date"},
	})
}

func New() *Group {
	return &Group{
		model: NewModel(),
	}
}

type createGroupSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type addViewerSignals struct {
	TabID      string           `json:"tab_id"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	FormData   struct {
		Email string `json:"email" validate:"required,email,max=320"`
		Role  string `json:"role" validate:"required,oneof=viewer admin"`
	} `json:"formData"`
}

type updateGroupSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	FormData   struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type deleteGroupSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type tabSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type paymentPaidAtSignals struct {
	TabID        string `json:"tab_id"`
	PaidAtDialog struct {
		Value string `json:"value"`
	} `json:"paidAtDialog"`
}

// NewGroupPage shows the form to create a new group
// EditGroupPage shows the form to edit a group name.
// CreateGroup handles group creation
func (g *Group) CreateGroup(c echo.Context) error {
	userID := utils.GetUserID(c)
	if userID == "" {
		err := utils.SSEHub.Redirect(c, "/login")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	signals := createGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = strings.TrimSpace(signals.FormData.Name)
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.Notify(c, errs["name"])
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	name := signals.FormData.Name
	// Billing gate is temporarily disabled.

	// Create group
	group, err := groupstore.CreateGroup(c.Request().Context(), groupstore.CreateGroupParams{
		ID:          utils.GenerateID("grp"),
		Name:        name,
		AdminUserID: userID,
	})
	if err != nil {
		slog.Error("group: failed to create group", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to create group")
	}

	slog.Info("group: created", "group_id", group.ID, "name", group.Name, "admin", userID)
	if userEmail := getUserEmail(c); userEmail != "" {
		err = email.Email().SendGroupCreated(c.Request().Context(), userEmail, group.Name, group.ID, utils.Env().URL)
		if err != nil {
			slog.Warn("group.create: failed to send group created email", "group_id", group.ID, "err", err)
		}
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.created"))

	// Redirect to group events
	err = utils.SSEHub.Redirect(c, "/groups/"+group.ID+"/events")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// IndexPage lists groups the user can access.
// RootPage redirects /groups/:groupId to the events tab.
func (g *Group) TogglePaymentEventPaid(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil && signals.TabID != "" {
		if !utils.SetTabID(c, signals.TabID) {
			return c.NoContent(http.StatusBadRequest)
		}
	} else {
		utils.EnsureTabID(c)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		return c.NoContent(http.StatusBadRequest)
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, targetPaidForTogglePage(pageKind), paymentsFadeRowKeyForEvent(eventID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedEvent, err := eventstore.ToggleEventPaid(c.Request().Context(), eventstore.ToggleEventPaidParams{ID: eventID, GroupID: groupID})
	if err != nil {
		slog.Error("group.payments.toggle_event_paid: failed", "group_id", groupID, "event_id", eventID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	notifyPaidToggleResult(c, updatedEvent.Paid)
	utils.InvalidateGroupCaches(groupID)
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (g *Group) UpdatePaymentEventPaidAt(c echo.Context) error {
	signals := paymentPaidAtSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	paidAt := normalizePaymentsPaidAtInput(signals.PaidAtDialog.Value)
	targetPaid := int64(0)
	if paidAt != nil {
		targetPaid = 1
	}
	eventID := c.Param("id")
	if !utils.IsValidID(eventID, utils.PrefixEvent) {
		return c.NoContent(http.StatusBadRequest)
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, targetPaid, paymentsFadeRowKeyForEvent(eventID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedEvent, err := eventstore.UpdateEventPaidAt(c.Request().Context(), eventstore.UpdateEventPaidAtParams{
		PaidAt:  paidAt,
		ID:      eventID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group.payments.update_event_paid_at: failed", "group_id", groupID, "event_id", eventID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "events.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)
	_ = updatedEvent
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	patchPaymentsPaidAtDialogClosed(c)
	return c.NoContent(http.StatusOK)
}

func (g *Group) TogglePaymentParticipantPaid(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil && signals.TabID != "" {
		if !utils.SetTabID(c, signals.TabID) {
			return c.NoContent(http.StatusBadRequest)
		}
	} else {
		utils.EnsureTabID(c)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	eventID := c.Param("eventId")
	memberID := c.Param("memberId")
	if !utils.IsValidID(eventID, utils.PrefixEvent) || !utils.IsValidID(memberID, utils.PrefixMember) {
		return c.NoContent(http.StatusBadRequest)
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, targetPaidForTogglePage(pageKind), paymentsFadeRowKeyForParticipant(eventID, memberID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedParticipant, err := eventstore.ToggleParticipantPaid(c.Request().Context(), eventstore.ToggleParticipantPaidParams{
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("group.payments.toggle_participant_paid: failed", "group_id", groupID, "event_id", eventID, "member_id", memberID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	notifyPaidToggleResult(c, updatedParticipant.Paid)
	utils.InvalidateGroupCaches(groupID)
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (g *Group) UpdatePaymentParticipantPaidAt(c echo.Context) error {
	signals := paymentPaidAtSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	paidAt := normalizePaymentsPaidAtInput(signals.PaidAtDialog.Value)
	targetPaid := int64(0)
	if paidAt != nil {
		targetPaid = 1
	}
	eventID := c.Param("eventId")
	memberID := c.Param("memberId")
	if !utils.IsValidID(eventID, utils.PrefixEvent) || !utils.IsValidID(memberID, utils.PrefixMember) {
		return c.NoContent(http.StatusBadRequest)
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, targetPaid, paymentsFadeRowKeyForParticipant(eventID, memberID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedParticipant, err := eventstore.UpdateParticipantPaidAt(c.Request().Context(), eventstore.UpdateParticipantPaidAtParams{
		PaidAt:   paidAt,
		EventID:  eventID,
		MemberID: memberID,
		GroupID:  groupID,
	})
	if err != nil {
		slog.Error("group.payments.update_participant_paid_at: failed", "group_id", groupID, "event_id", eventID, "member_id", memberID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "participants.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)
	_ = updatedParticipant
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	patchPaymentsPaidAtDialogClosed(c)
	return c.NoContent(http.StatusOK)
}

func (g *Group) TogglePaymentExpensePaid(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil && signals.TabID != "" {
		if !utils.SetTabID(c, signals.TabID) {
			return c.NoContent(http.StatusBadRequest)
		}
	} else {
		utils.EnsureTabID(c)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	expenseID := c.Param("id")
	if !utils.IsValidID(expenseID, utils.PrefixExpense) {
		return c.NoContent(http.StatusBadRequest)
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, targetPaidForTogglePage(pageKind), paymentsFadeRowKeyForExpense(expenseID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedExpense, err := expensestore.ToggleExpensePaid(c.Request().Context(), expensestore.ToggleExpensePaidParams{ID: expenseID, GroupID: groupID})
	if err != nil {
		slog.Error("group.payments.toggle_expense_paid: failed", "group_id", groupID, "expense_id", expenseID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	notifyPaidToggleResult(c, updatedExpense.Paid)
	utils.InvalidateGroupCaches(groupID)
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func (g *Group) UpdatePaymentExpensePaidAt(c echo.Context) error {
	signals := paymentPaidAtSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	groupID := utils.GetGroupID(c)
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	expenseID := c.Param("id")
	if !utils.IsValidID(expenseID, utils.PrefixExpense) {
		return c.NoContent(http.StatusBadRequest)
	}
	expense, err := expensestore.GetExpense(c.Request().Context(), expensestore.GetExpenseParams{
		ID:      expenseID,
		GroupID: groupID,
	})
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	paidAtValue := normalizePaymentsPaidAtInput(signals.PaidAtDialog.Value)
	paid := expense.Paid
	if paid == 0 && paidAtValue != nil {
		paid = 1
	}
	fadeHTML, shouldApplyFade, err := g.preparePaymentsRowFadeHTML(c, groupID, pageKind, paid, paymentsFadeRowKeyForExpense(expenseID))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	updatedExpense, err := expensestore.UpdateExpense(c.Request().Context(), expensestore.UpdateExpenseParams{
		Title:       expense.Title,
		Description: expense.Description,
		Amount:      expense.Amount,
		Date:        expense.Date,
		Paid:        paid,
		PaidAt:      paidAtValue,
		ID:          expenseID,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("group.payments.update_expense_paid_at: failed", "group_id", groupID, "expense_id", expenseID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)
	_ = updatedExpense
	if shouldApplyFade {
		utils.SSEHub.PatchHTML(c, fadeHTML)
		time.Sleep(paymentsFadeDuration)
	}
	if err := g.patchCurrentPaymentsPage(c, groupID); err != nil {
		return err
	}
	patchPaymentsPaidAtDialogClosed(c)
	return c.NoContent(http.StatusOK)
}

// UpdateGroup updates group name (admin only).
func (g *Group) UpdateGroup(c echo.Context) error {
	groupID := utils.GetGroupID(c)

	signals := updateGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors([]string{"name"}, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err := groupstore.UpdateGroupName(c.Request().Context(), groupstore.UpdateGroupNameParams{
		Name: signals.FormData.Name,
		ID:   groupID,
	})
	if err != nil {
		slog.Error("group.update: failed to update group", "group_id", groupID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.updated"))

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/about")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// LeaveGroup removes viewer access for the current user.
func (g *Group) LeaveGroup(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := utils.GetGroupID(c)
	userID := utils.GetUserID(c)
	var err error
	if userID == "" {
		err = utils.SSEHub.Redirect(c, "/login")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	_, err = groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.group_not_found"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if isAdminUser(c.Request().Context(), groupID, userID) {
		if err := g.removeAdminAccess(c.Request().Context(), groupID, userID); err != nil {
			if err == errAtLeastOneAdmin {
				utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.at_least_one_admin"))
			} else {
				slog.Error("group: failed to leave as admin", "group_id", groupID, "user_id", userID, "err", err)
				utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.leave_failed"))
			}
			err = utils.SSEHub.Redirect(c, "/groups")
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			return c.NoContent(http.StatusOK)
		}

		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.left"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	err = groupstore.RemoveGroupReader(c.Request().Context(), groupstore.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to leave", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.leave_failed"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.left"))
	err = utils.SSEHub.Redirect(c, "/groups")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// DeleteGroup removes the group and all data (admin only).
func (g *Group) DeleteGroup(c echo.Context) error {
	groupID := utils.GetGroupID(c)
	userID := utils.GetUserID(c)
	signals := deleteGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	var err error

	_, err = groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.group_not_found"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if err := groupstore.DeleteGroup(c.Request().Context(), groupID); err != nil {
		slog.Error("group: failed to delete", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.errors.delete_failed"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.deleted"))
	if signals.Mode == "table" {
		query := utils.NormalizeTableQuery(signals.TableQuery, g.model.TableQuerySpec())
		data, err := g.model.GetGroupsPageData(c.Request().Context(), userID, query)
		if err != nil {
			slog.Error("group.delete: failed to load dashboard data", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		data.Title = ctxi18n.T(c.Request().Context(), "groups.page_title")
		data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}}
		data.Signals = map[string]any{"mode": "table", "tableQuery": utils.TableQuerySignals(data.Query)}
		data.IsAuthenticated = true
		data.IsSuperAdmin = utils.IsSuperadmin(c)

		html, err := utils.RenderHTMLForRequest(c, GroupIndexPage(data))
		if err != nil {
			slog.Error("group.delete: failed to render dashboard", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	err = utils.SSEHub.Redirect(c, "/groups")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// UsersPage shows unified users table and invite form.
// UsersEntryPage shows a user or invite row details based on ID prefix.
// UsersNewPage shows the form to invite a new viewer/admin.
// UserEditPage shows user role edit page.
// UserPage shows details for a user access row.
// UserInvitePage shows details for a pending invite row.
func (g *Group) redirectUsersPage(c echo.Context, groupID, messageKey, errorKey string, status int) error {
	if messageKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), errorKey))
	}
	err := utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(status)
}

func (g *Group) patchUsersPage(c echo.Context, groupID, messageKey, errorKey string) error {
	if messageKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), errorKey))
	}
	data, err := g.usersPageData(c, groupID, queryValuesFromReferer(c))
	if err != nil {
		slog.Error("group: failed to load users patch data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	html, err := utils.RenderHTMLForRequest(c, GroupUsersPage(data))
	if err != nil {
		slog.Error("group: failed to render users page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func (g *Group) patchUsersPageWithState(c echo.Context, groupID string, query utils.TableQuery, messageKey, errorKey string) error {
	if messageKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), errorKey))
	}
	data, err := g.usersPageData(c, groupID, tableQueryValues(query))
	if err != nil {
		slog.Error("group: failed to load users patch data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	html, err := utils.RenderHTMLForRequest(c, GroupUsersPage(data))
	if err != nil {
		slog.Error("group: failed to render users page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

// AddViewer invites a user to the group with selected user role.
func (g *Group) AddViewer(c echo.Context) error {
	groupID := utils.GetGroupID(c)
	signals := addViewerSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Email = strings.ToLower(strings.TrimSpace(signals.FormData.Email))
	signals.FormData.Role = normalizeInviteRole(signals.FormData.Role)
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.Notify(c, errs["email"])
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	emailAddress := signals.FormData.Email
	inviteRole := signals.FormData.Role
	var err error

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.group_not_found")
	}

	// If user exists and already has access, short-circuit
	user, err := authstore.GetUserByEmail(c.Request().Context(), emailAddress)
	if err == nil {
		userRole, roleErr := getGroupAccessRole(c.Request().Context(), groupID, user.ID)
		if roleErr == nil && (userRole == "owner" || userRole == "admin") {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_admin", "")
		}
		if roleErr == nil && userRole == "viewer" {
			if inviteRole == "admin" {
				err = groupstore.RemoveGroupReader(c.Request().Context(), groupstore.RemoveGroupReaderParams{
					UserID:  user.ID,
					GroupID: groupID,
				})
				if err == nil {
					_, err = groupstore.CreateGroupAdmin(c.Request().Context(), groupstore.CreateGroupAdminParams{
						ID:      utils.GenerateID("gad"),
						UserID:  user.ID,
						GroupID: groupID,
					})
				}
				if err != nil {
					slog.Error("group: failed to promote viewer to admin", "group_id", groupID, "user_id", user.ID, "err", err)
					return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
				}
				if err := sendRoleChangeEmail(c.Request().Context(), user, group.Name, group.ID, "admin"); err != nil {
					slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", user.ID, "err", err)
				}
				return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.viewer_promoted", "")
			}
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_viewer", "")
		}
	}

	// Create invite magic link that does not expire.
	token := utils.GenerateID("tok")

	_, err = groupstore.CreateInviteMagicLink(c.Request().Context(), groupstore.CreateInviteMagicLinkParams{
		ID:         utils.GenerateID("mag"),
		Token:      token,
		Email:      emailAddress,
		GroupID:    sql.NullString{String: groupID, Valid: true},
		InviteRole: inviteRole,
	})
	if err != nil {
		slog.Error("group: failed to create invite link", "err", err)
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invite_failed")
	}

	err = email.Email().SendGroupInvitation(c.Request().Context(), emailAddress, group.Name, token, utils.Env().URL)
	if err != nil {
		slog.Error("group: failed to send invite email", "err", err)
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.send_failed")
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.invite_sent"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// RemoveViewer removes user access from the group.
func (g *Group) RemoveViewer(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := utils.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	if !utils.IsValidID(userID, "usr") {
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	currentUserID := utils.GetUserID(c)
	if currentUserID == userID {
		if utils.IsOwner(c) {
			return g.redirectUsersPage(c, groupID, "", "groups.errors.owner_cannot_leave", http.StatusConflict)
		}
	}
	if isAdminUser(ctx, groupID, userID) {
		if err := g.removeAdminAccess(ctx, groupID, userID); err != nil {
			if err == errAtLeastOneAdmin {
				return g.redirectUsersPage(c, groupID, "", "groups.errors.at_least_one_admin", http.StatusConflict)
			}
			slog.Error("group: failed to remove admin access", "group_id", groupID, "user_id", userID, "err", err)
			return g.redirectUsersPage(c, groupID, "", "groups.errors.remove_failed", http.StatusInternalServerError)
		}
		notifyAccessRemoved(ctx, groupID, userID)
		return g.redirectUsersPage(c, groupID, "groups.messages.viewer_removed", "", http.StatusOK)
	}

	err := groupstore.RemoveGroupReader(ctx, groupstore.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to remove viewer", "err", err)
		return g.redirectUsersPage(c, groupID, "", "groups.errors.remove_failed", http.StatusInternalServerError)
	}
	notifyAccessRemoved(ctx, groupID, userID)

	return g.redirectUsersPage(c, groupID, "groups.messages.viewer_removed", "", http.StatusOK)
}

func (g *Group) PromoteViewerToAdmin(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := utils.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	ctx := c.Request().Context()
	if !utils.IsValidID(userID, "usr") {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invalid_user")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	if isAdminUser(ctx, groupID, userID) {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_admin", "")
		}
		return g.redirectUsersPage(c, groupID, "groups.messages.already_admin", "", http.StatusOK)
	}

	role, roleErr := getGroupAccessRole(ctx, groupID, userID)
	if roleErr != nil || role != "viewer" {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	err := groupstore.RemoveGroupReader(ctx, groupstore.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to remove viewer while promoting", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	_, err = groupstore.CreateGroupAdmin(ctx, groupstore.CreateGroupAdminParams{
		ID:      utils.GenerateID("gad"),
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to promote viewer", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err == nil {
		if user, userErr := authstore.GetUserByID(ctx, userID); userErr == nil {
			if mailErr := sendRoleChangeEmail(ctx, user, group.Name, group.ID, "admin"); mailErr != nil {
				slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", userID, "err", mailErr)
			}
		}
	}

	if signals.Mode == "table" {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.viewer_promoted", "")
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.viewer_promoted"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users/"+userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (g *Group) DemoteAdminToViewer(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := utils.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	currentUserID := utils.GetUserID(c)
	ctx := c.Request().Context()
	if !utils.IsValidID(userID, "usr") {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invalid_user")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	if !isAdminUser(ctx, groupID, userID) {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.demote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.demote_failed", http.StatusInternalServerError)
	}
	if err := g.demoteAdminToViewer(ctx, groupID, userID); err != nil {
		if err == errAtLeastOneAdmin {
			if signals.Mode == "table" {
				return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.at_least_one_admin")
			}
			return g.redirectUsersPage(c, groupID, "", "groups.errors.at_least_one_admin", http.StatusConflict)
		}
		slog.Error("group: failed to demote admin", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.demote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.demote_failed", http.StatusInternalServerError)
	}
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err == nil {
		if user, userErr := authstore.GetUserByID(ctx, userID); userErr == nil {
			if mailErr := sendRoleChangeEmail(ctx, user, group.Name, group.ID, "viewer"); mailErr != nil {
				slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", userID, "err", mailErr)
			}
		}
	}

	if currentUserID == userID {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.admin_demoted"))
		tabID := utils.TabIDFromContext(c.Request().Context())
		redirectURL := "/groups/" + groupID + "/users"
		if tabID != "" {
			redirectURL += "?tab_id=" + tabID
		}
		err = utils.SSEHub.Redirect(c, redirectURL)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	if signals.Mode == "table" {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.admin_demoted", "")
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "groups.messages.admin_demoted"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users/"+userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// CancelInvite removes a pending invitation from the group.
func (g *Group) CancelInvite(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := utils.GetGroupID(c)
	inviteID := c.Param("inviteId")
	if inviteID == "" {
		inviteID = c.Param("id")
	}
	if !utils.IsValidID(inviteID, "mag") {
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_invite", http.StatusBadRequest)
	}

	err := groupstore.DeleteGroupPendingInvite(c.Request().Context(), groupstore.DeleteGroupPendingInviteParams{
		ID:      inviteID,
		GroupID: sql.NullString{String: groupID, Valid: true},
	})
	if err != nil {
		slog.Error("group: failed to cancel invite", "err", err)
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invite_cancel_failed", http.StatusInternalServerError)
	}

	return g.redirectUsersPage(c, groupID, "groups.messages.invite_cancelled", "", http.StatusOK)
}

func (g *Group) DeleteUserEntry(c echo.Context) error {
	id := c.Param("id")
	if utils.IsValidID(id, "usr") {
		return g.RemoveViewer(c)
	}
	if utils.IsValidID(id, "mag") {
		return g.CancelInvite(c)
	}
	return c.NoContent(http.StatusBadRequest)
}

func (g *Group) usersPageData(c echo.Context, groupID string, values url.Values) (UsersPageData, error) {
	ctx := c.Request().Context()
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return UsersPageData{}, err
	}
	query := parseTableQueryFromValues(values, usersTableQuerySpec())

	rows, err := g.buildUserRows(ctx, groupID)
	if err != nil {
		return UsersPageData{}, err
	}

	rows = filterUserRows(rows, query.Search)
	sortUserRows(rows, query)
	total := int64(len(rows))
	query = utils.ClampPage(query, total)
	rows = pageUserRows(rows, query)

	return UsersPageData{
		Title: ctxi18n.T(ctx, "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + group.ID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.users"), Href: "/groups/" + group.ID + "/users"},
		},
		CurrentUserID:   utils.GetUserID(c),
		Group:           group,
		UserRows:        rows,
		IsAdmin:         utils.IsAdmin(c),
		Query:           query,
		Pager:           utils.BuildTablePagination(total, query),
		GroupID:         groupID,
		UsersTable:      GroupUsersTableLayout(),
		Signals:         map[string]any{"mode": "table", "tableQuery": utils.TableQuerySignals(query)},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}, nil
}

func queryValuesFromReferer(c echo.Context) url.Values {
	referer := c.Request().Referer()
	if referer == "" {
		return url.Values{}
	}
	u, err := url.Parse(referer)
	if err != nil {
		return url.Values{}
	}
	return u.Query()
}

func tableQueryValues(query utils.TableQuery) url.Values {
	values := url.Values{}
	values.Set("page", strconv.Itoa(query.Page))
	values.Set("pageSize", strconv.Itoa(query.PageSize))
	values.Set("q", query.Search)
	if query.SortSet && query.Sort != "" {
		values.Set("sort", query.Sort)
	}
	if query.Dir != "" {
		values.Set("dir", query.Dir)
	}
	return values
}

func (g *Group) buildUserRows(ctx context.Context, groupID string) ([]GroupUserRow, error) {
	userAccessRows, err := groupstore.ListGroupUserAccess(ctx, groupID)
	if err != nil {
		return nil, err
	}
	rows := make([]GroupUserRow, 0, len(userAccessRows))
	for _, accessRow := range userAccessRows {
		createdAt := time.Time{}
		if accessRow.AccessCreatedAt.Valid {
			createdAt = accessRow.AccessCreatedAt.Time
		} else if accessRow.CreatedAt.Valid {
			createdAt = accessRow.CreatedAt.Time
		}
		rows = append(rows, GroupUserRow{
			Kind:      "user",
			Status:    "active",
			Role:      accessRow.Role,
			Email:     accessRow.Email,
			UserID:    accessRow.ID,
			CreatedAt: createdAt,
		})
	}

	invites, err := groupstore.ListGroupPendingInvites(ctx, sql.NullString{String: groupID, Valid: true})
	if err != nil {
		return nil, err
	}
	for _, invite := range invites {
		createdAt := time.Time{}
		if invite.CreatedAt.Valid {
			createdAt = invite.CreatedAt.Time
		}
		rows = append(rows, GroupUserRow{
			Kind:      "invite",
			Status:    "pending",
			Role:      normalizeInviteRole(invite.InviteRole),
			Email:     invite.Email,
			InviteID:  invite.ID,
			CreatedAt: createdAt,
		})
	}

	return rows, nil
}

func filterUserRows(rows []GroupUserRow, search string) []GroupUserRow {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return rows
	}
	filtered := make([]GroupUserRow, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Email), search) || strings.Contains(strings.ToLower(row.Role), search) || strings.Contains(strings.ToLower(row.Status), search) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func sortUserRows(rows []GroupUserRow, query utils.TableQuery) {
	less := func(i, j int) bool {
		a := rows[i]
		b := rows[j]
		switch query.Sort {
		case "email":
			return strings.ToLower(a.Email) < strings.ToLower(b.Email)
		case "role":
			return a.Role < b.Role
		case "status":
			return a.Status < b.Status
		default:
			if a.CreatedAt.Equal(b.CreatedAt) {
				return strings.ToLower(a.Email) < strings.ToLower(b.Email)
			}
			return a.CreatedAt.Before(b.CreatedAt)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if query.Dir == "desc" {
			return !less(i, j)
		}
		return less(i, j)
	})
}

func pageUserRows(rows []GroupUserRow, query utils.TableQuery) []GroupUserRow {
	if len(rows) == 0 {
		return rows
	}
	start := int(query.Offset())
	if start >= len(rows) {
		return []GroupUserRow{}
	}
	end := start + query.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func listGroupAdmins(ctx context.Context, groupID string) ([]db.User, error) {
	return groupstore.ListGroupAdmins(ctx, groupID)
}

func isAdminUser(ctx context.Context, groupID, userID string) bool {
	role, err := getGroupAccessRole(ctx, groupID, userID)
	if err != nil {
		return false
	}
	return role == "owner" || role == "admin"
}

func getGroupAccessRole(ctx context.Context, groupID, userID string) (string, error) {
	return groupstore.GetGroupAccessRole(ctx, groupstore.GetGroupAccessRoleParams{UserID: userID, GroupID: groupID})
}

func (g *Group) demoteAdminToViewer(ctx context.Context, groupID, userID string) error {
	if err := g.removeAdminAccess(ctx, groupID, userID); err != nil {
		return err
	}
	count, err := groupstore.IsGroupReader(ctx, groupstore.IsGroupReaderParams{UserID: userID, GroupID: groupID})
	if err == nil && count > 0 {
		return nil
	}
	_, err = groupstore.CreateGroupReader(ctx, groupstore.CreateGroupReaderParams{ID: utils.GenerateID("grd"), UserID: userID, GroupID: groupID})
	return err
}

func (g *Group) removeAdminAccess(ctx context.Context, groupID, userID string) error {
	adminUsers, err := listGroupAdmins(ctx, groupID)
	if err != nil {
		return err
	}
	if len(adminUsers) <= 1 {
		return errAtLeastOneAdmin
	}
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.AdminUserID == userID {
		replacement := ""
		for _, admin := range adminUsers {
			if admin.ID != userID {
				replacement = admin.ID
				break
			}
		}
		if replacement == "" {
			return errAtLeastOneAdmin
		}
		if err := groupstore.UpdateGroupAdmin(ctx, groupstore.UpdateGroupAdminParams{AdminUserID: replacement, ID: groupID}); err != nil {
			return err
		}
	}
	_ = groupstore.RemoveGroupAdmin(ctx, groupstore.RemoveGroupAdminParams{UserID: userID, GroupID: groupID})
	_ = groupstore.RemoveGroupReader(ctx, groupstore.RemoveGroupReaderParams{UserID: userID, GroupID: groupID})
	return nil
}

func sendRoleChangeEmail(ctx context.Context, user db.User, groupName, groupID, role string) error {
	baseURL := utils.Env().URL
	mailCtx := ctx
	if user.PreferredLang != "" {
		if localizedCtx, err := ctxi18nlib.WithLocale(ctx, user.PreferredLang); err == nil {
			mailCtx = localizedCtx
		}
	}
	if role == "admin" {
		return email.Email().SendRoleUpgradedToAdmin(mailCtx, user.Email, groupName, groupID, baseURL)
	}
	return email.Email().SendRoleDowngradedToViewer(mailCtx, user.Email, groupName, groupID, baseURL)
}

func notifyAccessRemoved(ctx context.Context, groupID, userID string) {
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Warn("group: failed to load group for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}
	user, err := authstore.GetUserByID(ctx, userID)
	if err != nil {
		slog.Warn("group: failed to load user for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}
	admins, err := listGroupAdmins(ctx, groupID)
	if err != nil {
		slog.Warn("group: failed to list admins for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}

	adminEmails := make([]string, 0, len(admins))
	for _, admin := range admins {
		adminEmails = append(adminEmails, admin.Email)
	}

	mailCtx := ctx
	if user.PreferredLang != "" {
		if localizedCtx, localeErr := ctxi18nlib.WithLocale(ctx, user.PreferredLang); localeErr == nil {
			mailCtx = localizedCtx
		}
	}

	if err := email.Email().SendAccessRemoved(mailCtx, user.Email, group.Name, adminEmails, utils.Env().URL); err != nil {
		slog.Warn("group: failed to send access-removed email", "group_id", groupID, "user_id", userID, "err", err)
	}
}

func parseTableQueryFromValues(values url.Values, spec utils.TableQuerySpec) utils.TableQuery {
	query := utils.TableQuery{
		Page:     parsePositiveInt(values.Get("page"), 1),
		PageSize: parsePositiveInt(values.Get("pageSize"), utils.DefaultTablePageSize),
		Search:   strings.TrimSpace(values.Get("q")),
		Sort:     values.Get("sort"),
		Dir:      values.Get("dir"),
		SortSet:  values.Get("sort") != "",
	}

	return utils.NormalizeTableQuery(query, spec)
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func dateMatchesTableQuery(date string, query utils.TableQuery) bool {
	if query.From != "" && query.To != "" {
		return date >= query.From && date <= query.To
	}
	if query.Year != "" {
		return strings.HasPrefix(date, query.Year+"-")
	}
	return true
}

func filterPaymentEventRows(rows []GroupPaymentEventRow, query utils.TableQuery) []GroupPaymentEventRow {
	search := strings.ToLower(strings.TrimSpace(query.Search))
	if search == "" && query.From == "" && query.To == "" && query.Year == "" {
		return rows
	}
	filtered := make([]GroupPaymentEventRow, 0, len(rows))
	for _, row := range rows {
		if search != "" {
			if !strings.Contains(strings.ToLower(row.Title), search) {
				continue
			}
		}
		if !dateMatchesTableQuery(row.Date, query) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func sortPaymentEventRows(rows []GroupPaymentEventRow, query utils.TableQuery) {
	less := func(a, b GroupPaymentEventRow) bool {
		switch query.Sort {
		case "title":
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "amount":
			return a.Amount < b.Amount
		case "paid_at":
			return a.PaidAt < b.PaidAt
		case "paid":
			return (a.PaidAt != "") && (b.PaidAt == "")
		default:
			return a.Date < b.Date
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if query.Dir == "desc" {
			return !less(rows[i], rows[j])
		}
		return less(rows[i], rows[j])
	})
}

func pagePaymentEventRows(rows []GroupPaymentEventRow, query utils.TableQuery) []GroupPaymentEventRow {
	if len(rows) == 0 {
		return rows
	}
	start := int(query.Offset())
	if start >= len(rows) {
		return []GroupPaymentEventRow{}
	}
	end := start + query.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func filterOutgoingPaymentRows(rows []GroupOutgoingPaymentRow, query utils.TableQuery) []GroupOutgoingPaymentRow {
	search := strings.ToLower(strings.TrimSpace(query.Search))
	if search == "" && query.From == "" && query.To == "" && query.Year == "" {
		return rows
	}
	filtered := make([]GroupOutgoingPaymentRow, 0, len(rows))
	for _, row := range rows {
		if search != "" {
			match := strings.Contains(strings.ToLower(row.Title), search) ||
				strings.Contains(strings.ToLower(row.EventTitle), search) ||
				strings.Contains(strings.ToLower(row.MemberName), search) ||
				strings.Contains(strings.ToLower(row.Kind), search)
			if !match {
				continue
			}
		}
		if !dateMatchesTableQuery(row.Date, query) {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func sortOutgoingPaymentRows(rows []GroupOutgoingPaymentRow, query utils.TableQuery) {
	less := func(a, b GroupOutgoingPaymentRow) bool {
		switch query.Sort {
		case "type":
			return a.Kind < b.Kind
		case "title":
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "amount":
			return a.Amount < b.Amount
		case "paid_at":
			return a.PaidAt < b.PaidAt
		case "paid":
			return (a.PaidAt != "") && (b.PaidAt == "")
		default:
			return a.Date < b.Date
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if query.Dir == "desc" {
			return !less(rows[i], rows[j])
		}
		return less(rows[i], rows[j])
	})
}

func pageOutgoingPaymentRows(rows []GroupOutgoingPaymentRow, query utils.TableQuery) []GroupOutgoingPaymentRow {
	if len(rows) == 0 {
		return rows
	}
	start := int(query.Offset())
	if start >= len(rows) {
		return []GroupOutgoingPaymentRow{}
	}
	end := start + query.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func normalizeInviteRole(role string) string {
	if strings.EqualFold(strings.TrimSpace(role), "admin") {
		return "admin"
	}
	return "viewer"
}

type paymentsPageKind string

const (
	paymentsPageToReceive      paymentsPageKind = "pending-incomes"
	paymentsPageToPay          paymentsPageKind = "pending-payouts"
	paymentsPageRecentIncome   paymentsPageKind = "recent-incomes"
	paymentsPageRecentOutgoing paymentsPageKind = "recent-payouts"
)

func detectPaymentsPageFromReferer(c echo.Context, groupID string) paymentsPageKind {
	path := ""
	referer := c.Request().Referer()
	if referer != "" {
		if parsed, err := url.Parse(referer); err == nil {
			path = parsed.Path
		}
	}
	base := "/groups/" + groupID + "/"
	switch path {
	case base + "to-receive", base + "pending-incomes":
		return paymentsPageToReceive
	case base + "recent-income", base + "recent-incomes":
		return paymentsPageRecentIncome
	case base + "recent-outgoing", base + "recent-payouts":
		return paymentsPageRecentOutgoing
	case base + "to-pay":
		return paymentsPageToPay
	default:
		return paymentsPageToPay
	}
}

func (g *Group) renderPaymentsPageHTML(c echo.Context, groupID string, pageKind paymentsPageKind, fadeRowKey string) (string, error) {
	values := queryValuesFromReferer(c)
	switch pageKind {
	case paymentsPageToReceive:
		query := parseTableQueryFromValues(values, toReceivePaymentsTableQuerySpec())
		data, err := g.toReceivePageData(c, groupID, query)
		if err != nil {
			return "", err
		}
		data.FadeRowKey = fadeRowKey
		return utils.RenderHTMLForRequest(c, GroupToReceivePage(data))
	case paymentsPageRecentIncome:
		query := parseTableQueryFromValues(values, recentIncomePaymentsTableQuerySpec())
		data, err := g.recentIncomePageData(c, groupID, query)
		if err != nil {
			return "", err
		}
		data.FadeRowKey = fadeRowKey
		return utils.RenderHTMLForRequest(c, GroupRecentIncomePage(data))
	case paymentsPageRecentOutgoing:
		query := parseTableQueryFromValues(values, recentOutgoingPaymentsTableQuerySpec())
		data, err := g.recentOutgoingPageData(c, groupID, query)
		if err != nil {
			return "", err
		}
		data.FadeRowKey = fadeRowKey
		return utils.RenderHTMLForRequest(c, GroupRecentOutgoingPage(data))
	default:
		query := parseTableQueryFromValues(values, toPayPaymentsTableQuerySpec())
		data, err := g.toPayPageData(c, groupID, query)
		if err != nil {
			return "", err
		}
		data.FadeRowKey = fadeRowKey
		return utils.RenderHTMLForRequest(c, GroupToPayPage(data))
	}
}

func (g *Group) patchCurrentPaymentsPage(c echo.Context, groupID string) error {
	pageKind := detectPaymentsPageFromReferer(c, groupID)
	html, err := g.renderPaymentsPageHTML(c, groupID, pageKind, "")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)
	return nil
}

func shouldFadePaymentsRowRemoval(pageKind paymentsPageKind, paid int64) bool {
	switch pageKind {
	case paymentsPageToReceive, paymentsPageToPay:
		return paid == 1
	case paymentsPageRecentIncome, paymentsPageRecentOutgoing:
		return paid == 0
	default:
		return false
	}
}

func targetPaidForTogglePage(pageKind paymentsPageKind) int64 {
	switch pageKind {
	case paymentsPageToReceive, paymentsPageToPay:
		return 1
	case paymentsPageRecentIncome, paymentsPageRecentOutgoing:
		return 0
	default:
		return -1
	}
}

func (g *Group) preparePaymentsRowFadeHTML(c echo.Context, groupID string, pageKind paymentsPageKind, targetPaid int64, fadeRowKey string) (string, bool, error) {
	if strings.TrimSpace(fadeRowKey) == "" {
		return "", false, nil
	}
	if !shouldFadePaymentsRowRemoval(pageKind, targetPaid) {
		return "", false, nil
	}
	html, err := g.renderPaymentsPageHTML(c, groupID, pageKind, fadeRowKey)
	if err != nil {
		return "", false, err
	}
	return html, true, nil
}

func paymentsFadeRowKeyForEvent(eventID string) string {
	return "event:" + eventID
}

func paymentsFadeRowKeyForParticipant(eventID, memberID string) string {
	return "participant:" + eventID + ":" + memberID
}

func paymentsFadeRowKeyForExpense(expenseID string) string {
	return "expense:" + expenseID
}

func patchPaymentsPaidAtDialogClosed(c echo.Context) {
	utils.SSEHub.PatchSignals(c, map[string]any{"paidAtDialog": map[string]any{"open": false, "fetching": false}})
}

func notifyPaidToggleResult(c echo.Context, paid int64) {
	if paid == 1 {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "paid_status.marked_as_paid"))
		return
	}
	utils.Notify(c, ctxi18n.T(c.Request().Context(), "paid_status.marked_as_unpaid"))
}

func paymentsPaidAtFromNullString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return utils.FormatDateInput(value.String)
}

func normalizePaymentsPaidAtInput(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	formatted := utils.FormatDateInput(trimmed)
	if formatted == "" {
		return nil
	}
	return formatted
}

func getUserEmail(c echo.Context) string {
	userID := utils.GetUserID(c)
	if userID == "" {
		return ""
	}
	user, err := authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return ""
	}
	return user.Email
}

func (g *Group) groupPageData(c echo.Context, groupID string) (GroupPageData, error) {
	ctx := c.Request().Context()
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupPageData{}, err
	}

	admin, err := authstore.GetUserByID(ctx, group.AdminUserID)
	if err != nil {
		return GroupPageData{}, err
	}

	totals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return GroupPageData{}, err
	}

	return GroupPageData{
		Title:           "Bandcash - " + group.Name,
		Breadcrumbs:     []utils.Crumb{{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"}, {Label: group.Name, Href: "/groups/" + groupID + "/about"}, {Label: ctxi18n.T(ctx, "groups.about")}},
		Signals:         map[string]any{"mode": "single", "formState": "", "eventFormState": "", "summaryMode": "all", "formData": map[string]any{"name": group.Name}, "errors": map[string]any{"name": ""}},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
		Group:           group,
		Admin:           admin,
		Income:          totals.Income.All,
		IncomePaid:      totals.Income.Paid,
		IncomeUnpaid:    totals.Income.Unpaid,
		Payouts:         totals.Payouts.All,
		PayoutsPaid:     totals.Payouts.Paid,
		PayoutsUnpaid:   totals.Payouts.Unpaid,
		Expenses:        totals.Expenses.All,
		ExpensesPaid:    totals.Expenses.Paid,
		ExpensesUnpaid:  totals.Expenses.Unpaid,
		Balance:         totals.Balance.All,
		IsAdmin:         utils.IsAdmin(c),
	}, nil
}

func newPaymentsDialogSignals(ctx context.Context, query utils.TableQuery) map[string]any {
	return map[string]any{
		"mode":       "table",
		"tableQuery": utils.TableQuerySignals(query),
		"paidAtDialog": map[string]any{
			"open":        false,
			"fetching":    false,
			"title":       ctxi18n.T(ctx, "fields.paid_at"),
			"message":     "",
			"value":       "",
			"url":         "",
			"submitLabel": ctxi18n.T(ctx, "table.apply"),
			"cancelLabel": ctxi18n.T(ctx, "actions.cancel"),
		},
	}
}

func (g *Group) toReceivePageData(c echo.Context, groupID string, query utils.TableQuery) (GroupToReceivePageData, error) {
	ctx := c.Request().Context()

	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupToReceivePageData{}, err
	}
	rows, err := eventstore.ListUnpaidEventsByGroup(ctx, groupID)
	if err != nil {
		return GroupToReceivePageData{}, err
	}
	events := make([]GroupPaymentEventRow, 0, len(rows))
	for _, row := range rows {
		events = append(events, GroupPaymentEventRow{
			ID:     row.ID,
			Title:  row.Title,
			Amount: row.Amount,
			PaidAt: paymentsPaidAtFromNullString(row.PaidAt),
			Date:   utils.FormatDateInput(row.Time),
		})
	}
	events = filterPaymentEventRows(events, query)
	sortPaymentEventRows(events, query)
	total := int64(len(events))
	query = utils.ClampPage(query, total)
	events = pagePaymentEventRows(events, query)

	return GroupToReceivePageData{
		Title: ctxi18n.T(ctx, "groups.pending_incomes_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.pending_incomes")},
		},
		Signals:         newPaymentsDialogSignals(ctx, query),
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
		IsAdmin:         utils.IsAdmin(c),
		GroupID:         groupID,
		Group:           group,
		Rows:            events,
		EventsTable:     GroupPaymentsEventsTableLayout(),
		Query:           query,
		Pager:           utils.BuildTablePagination(total, query),
		RecentYears:     utils.RecentYears(3),
		FadeRowKey:      "",
	}, nil
}

func (g *Group) toPayPageData(c echo.Context, groupID string, query utils.TableQuery) (GroupToPayPageData, error) {
	ctx := c.Request().Context()

	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupToPayPageData{}, err
	}
	outgoingRows, err := groupstore.ListUnpaidOutgoingPaymentsByGroup(ctx, groupID)
	if err != nil {
		return GroupToPayPageData{}, err
	}
	rows := make([]GroupOutgoingPaymentRow, 0, len(outgoingRows))
	for _, row := range outgoingRows {
		rows = append(rows, GroupOutgoingPaymentRow{
			Kind:       row.PaymentKind,
			PaymentID:  row.PaymentID,
			Title:      row.Title,
			EventID:    row.EventID,
			EventTitle: row.EventTitle,
			MemberID:   row.MemberID,
			MemberName: row.MemberName,
			Amount:     row.Amount,
			PaidAt:     paymentsPaidAtFromNullString(row.PaidAt),
			Date:       utils.FormatDateInput(row.SortDate),
		})
	}
	rows = filterOutgoingPaymentRows(rows, query)
	sortOutgoingPaymentRows(rows, query)
	total := int64(len(rows))
	query = utils.ClampPage(query, total)
	rows = pageOutgoingPaymentRows(rows, query)

	return GroupToPayPageData{
		Title: ctxi18n.T(ctx, "groups.pending_payouts_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.pending_payouts")},
		},
		Signals:         newPaymentsDialogSignals(ctx, query),
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
		IsAdmin:         utils.IsAdmin(c),
		GroupID:         groupID,
		Group:           group,
		Rows:            rows,
		OutgoingTable:   GroupOutgoingPaymentsTableLayout(),
		Query:           query,
		Pager:           utils.BuildTablePagination(total, query),
		RecentYears:     utils.RecentYears(3),
		FadeRowKey:      "",
	}, nil
}

func (g *Group) recentIncomePageData(c echo.Context, groupID string, query utils.TableQuery) (GroupRecentIncomePageData, error) {
	ctx := c.Request().Context()

	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupRecentIncomePageData{}, err
	}
	rows, err := eventstore.ListPaidEventsByGroup(ctx, groupID)
	if err != nil {
		return GroupRecentIncomePageData{}, err
	}
	events := make([]GroupPaymentEventRow, 0, len(rows))
	for _, row := range rows {
		events = append(events, GroupPaymentEventRow{
			ID:     row.ID,
			Title:  row.Title,
			Amount: row.Amount,
			PaidAt: paymentsPaidAtFromNullString(row.PaidAt),
			Date:   utils.FormatDateInput(row.Time),
		})
	}
	events = filterPaymentEventRows(events, query)
	sortPaymentEventRows(events, query)
	total := int64(len(events))
	query = utils.ClampPage(query, total)
	events = pagePaymentEventRows(events, query)

	return GroupRecentIncomePageData{
		Title: ctxi18n.T(ctx, "groups.recent_incomes_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.recent_incomes")},
		},
		Signals:         newPaymentsDialogSignals(ctx, query),
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
		IsAdmin:         utils.IsAdmin(c),
		GroupID:         groupID,
		Group:           group,
		Rows:            events,
		EventsTable:     GroupPaymentsEventsTableLayout(),
		Query:           query,
		Pager:           utils.BuildTablePagination(total, query),
		RecentYears:     utils.RecentYears(3),
		FadeRowKey:      "",
	}, nil
}

func (g *Group) recentOutgoingPageData(c echo.Context, groupID string, query utils.TableQuery) (GroupRecentOutgoingPageData, error) {
	ctx := c.Request().Context()

	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupRecentOutgoingPageData{}, err
	}
	outgoingRows, err := groupstore.ListPaidOutgoingPaymentsByGroup(ctx, groupID)
	if err != nil {
		return GroupRecentOutgoingPageData{}, err
	}
	rows := make([]GroupOutgoingPaymentRow, 0, len(outgoingRows))
	for _, row := range outgoingRows {
		rows = append(rows, GroupOutgoingPaymentRow{
			Kind:       row.PaymentKind,
			PaymentID:  row.PaymentID,
			Title:      row.Title,
			EventID:    row.EventID,
			EventTitle: row.EventTitle,
			MemberID:   row.MemberID,
			MemberName: row.MemberName,
			Amount:     row.Amount,
			PaidAt:     paymentsPaidAtFromNullString(row.PaidAt),
			Date:       utils.FormatDateInput(row.SortDate),
		})
	}
	rows = filterOutgoingPaymentRows(rows, query)
	sortOutgoingPaymentRows(rows, query)
	total := int64(len(rows))
	query = utils.ClampPage(query, total)
	rows = pageOutgoingPaymentRows(rows, query)
	return GroupRecentOutgoingPageData{
		Title: ctxi18n.T(ctx, "groups.recent_payouts_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.recent_payouts")},
		},
		Signals:         newPaymentsDialogSignals(ctx, query),
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
		IsAdmin:         utils.IsAdmin(c),
		GroupID:         groupID,
		Group:           group,
		Rows:            rows,
		OutgoingTable:   GroupOutgoingPaymentsTableLayout(),
		Query:           query,
		Pager:           utils.BuildTablePagination(total, query),
		RecentYears:     utils.RecentYears(3),
		FadeRowKey:      "",
	}, nil
}
