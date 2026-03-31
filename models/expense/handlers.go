package expense

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

type expenseParams struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Date        string `json:"date" validate:"required"`
	Paid        bool   `json:"paid"`
	PaidAt      string `json:"paidAt"`
}

type expenseTableParams struct {
	TabID      string           `json:"tab_id"`
	FormData   expenseParams    `json:"formData"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	Mode       string           `json:"mode"`
}

type modeParams struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

var (
	defaultExpenseSignals = map[string]any{
		"mode":      "table",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false, "paidAt": ""},
		"errors":    map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
	expenseErrorFields = []string{"title", "description", "amount", "date"}
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

func applyExpenseTableByRole(data *ExpensesData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.ExpensesTable.ActionsWidthRem = 0
	}
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

func (e *Expenses) NewExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("expense.new_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := NewExpensePageData{
		Title: ctxi18n.T(c.Request().Context(), "expenses.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "expenses.title"), Href: "/groups/" + groupID + "/expenses"},
			{Label: ctxi18n.T(c.Request().Context(), "expenses.add")},
		},
		GroupID: groupID,
		Signals: map[string]any{
			"formData": map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false, "paidAt": ""},
			"errors":   map[string]any{"title": "", "description": "", "amount": "", "date": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    middleware.IsSuperadmin(c),
	}
	return utils.RenderPage(c, ExpenseNewPage(data))
}

func (e *Expenses) EditExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("expense.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	expense, err := db.Qry.GetExpense(c.Request().Context(), db.GetExpenseParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.edit_page: failed to get expense", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := EditExpensePageData{
		Title: ctxi18n.T(c.Request().Context(), "expenses.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "expenses.title"), Href: "/groups/" + groupID + "/expenses"},
			{Label: expense.Title, Href: "/groups/" + groupID + "/expenses/" + id},
			{Label: ctxi18n.T(c.Request().Context(), "expenses.edit")},
		},
		GroupID: groupID,
		Expense: &expense,
		Signals: map[string]any{
			"formData": map[string]any{
				"title":       expense.Title,
				"description": expense.Description,
				"amount":      expense.Amount,
				"date":        expense.Date,
				"paid":        expense.Paid == 1,
				"paidAt": func() string {
					if !expense.PaidAt.Valid {
						return ""
					}
					return utils.FormatDateInput(expense.PaidAt.String)
				}(),
			},
			"errors": map[string]any{"title": "", "description": "", "amount": "", "date": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    middleware.IsSuperadmin(c),
	}
	return utils.RenderPage(c, ExpenseEditPage(data))
}

func (e *Expenses) Index(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := utils.ParseTableQuery(c, e)

	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseIndex(data))
}

func (e *Expenses) Show(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := e.GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("expense.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.Signals = expenseShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseShow(data))
}

func (e *Expenses) Create(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	var signals expenseTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.Date = strings.TrimSpace(signals.FormData.Date)
	signals.FormData.PaidAt = normalizePaidAtInput(signals.FormData.PaidAt)

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(expenseErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err = db.Qry.CreateExpense(c.Request().Context(), db.CreateExpenseParams{
		ID:          utils.GenerateID(utils.PrefixExpense),
		GroupID:     groupID,
		Title:       signals.FormData.Title,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
		Date:        signals.FormData.Date,
		Paid: func() int64 {
			if signals.FormData.Paid {
				return 1
			}
			return 0
		}(),
		PaidAt: paidAtArg(signals.FormData.Paid, signals.FormData.PaidAt),
	})
	if err != nil {
		slog.Error("expense.create.table: failed to create expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.created"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/expenses")
	if err != nil {
		slog.Warn("expense.create: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func (e *Expenses) Update(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.update: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals expenseTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.update: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.Date = strings.TrimSpace(signals.FormData.Date)
	signals.FormData.PaidAt = normalizePaidAtInput(signals.FormData.PaidAt)

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors(expenseErrorFields, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err = db.Qry.UpdateExpense(c.Request().Context(), db.UpdateExpenseParams{
		Title:       signals.FormData.Title,
		Description: signals.FormData.Description,
		Amount:      signals.FormData.Amount,
		Date:        signals.FormData.Date,
		Paid: func() int64 {
			if signals.FormData.Paid {
				return 1
			}
			return 0
		}(),
		PaidAt:  paidAtArg(signals.FormData.Paid, signals.FormData.PaidAt),
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.update: failed to update expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/expenses/"+id)
	if err != nil {
		slog.Warn("expense.update: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func (e *Expenses) Destroy(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.destroy: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.destroy: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	err = db.Qry.DeleteExpense(c.Request().Context(), db.DeleteExpenseParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.destroy: failed to delete expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.deleted"))

	if signals.Mode == "single" {
		err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/expenses")
		if err != nil {
			slog.Warn("expense.destroy: failed to redirect", "err", err)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Expenses) TogglePaid(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.togglePaid: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals modeParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.togglePaid: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	result, err := db.Qry.ToggleExpensePaid(c.Request().Context(), db.ToggleExpensePaidParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.togglePaid: failed to toggle paid status", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("expense.togglePaid", "id", id)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	if result.Paid == 1 {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "paid_status.marked_as_paid"))
	} else {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "paid_status.marked_as_unpaid"))
	}

	if signals.Mode == "single" {
		data, err := e.GetShowData(c.Request().Context(), groupID, id)
		if err != nil {
			slog.Error("expense.togglePaid: failed to get data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		data.IsAdmin = middleware.IsAdmin(c)
		data.Signals = expenseShowSignals(data)
		data.IsAuthenticated = true
		data.IsSuperAdmin = middleware.IsSuperadmin(c)
		html, err := utils.RenderHTMLForRequest(c, ExpenseShow(data))
		if err != nil {
			slog.Error("expense.togglePaid: failed to render", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.togglePaid: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.togglePaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}
