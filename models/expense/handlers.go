package expense

import (
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

func NewExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	group, err := db.GetGroupByID(c.Request().Context(), groupID)
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

func EditExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	group, err := db.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("expense.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	expense, err := db.GetExpense(c.Request().Context(), db.GetExpenseParams{
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

func IndexPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: TableQuerySpec()})

	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseIndexPage(data))
}

func ShowPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("expense.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.Signals = expenseShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseShowPage(data))
}

func Create(c echo.Context) error {
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

	_, err = db.CreateExpense(c.Request().Context(), db.CreateExpenseParams{
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
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.created"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/expenses")
	if err != nil {
		slog.Warn("expense.create: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func Update(c echo.Context) error {
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

	_, err = db.UpdateExpense(c.Request().Context(), db.UpdateExpenseParams{
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
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/expenses/"+id)
	if err != nil {
		slog.Warn("expense.update: failed to redirect", "err", err)
	}
	return c.NoContent(http.StatusOK)
}

func Destroy(c echo.Context) error {
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

	err = db.DeleteExpense(c.Request().Context(), db.DeleteExpenseParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.destroy: failed to delete expense", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.delete_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.deleted"))

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

	query := utils.NormalizeTableQuery(signals.TableQuery, TableQuerySpec())
	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndexPage(data))
	if err != nil {
		slog.Error("expense.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func openPaidAtDialog(c echo.Context, groupID, id string, _ utils.TableQuery) error {
	data, err := GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("expense.openPaidAtDialog: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	data.PaidAtDialog = PaidAtDialogState{
		Open:  true,
		Mode:  "single",
		Title: ctxi18n.T(c.Request().Context(), "fields.paid_at"),
		Value: func() string {
			if data.Expense.PaidAt.Valid {
				return utils.FormatDateInput(data.Expense.PaidAt.String)
			}
			return ""
		}(),
		SubmitLabel: ctxi18n.T(c.Request().Context(), "table.apply"),
		CancelLabel: ctxi18n.T(c.Request().Context(), "actions.cancel"),
		URL:         "/groups/" + groupID + "/expenses/" + id + "/paid_at",
		TriggerID:   "expense-amount-paid-at-edit",
	}
	data.Signals = expenseShowSignals(data)

	html, err := utils.RenderHTMLForRequest(c, ExpenseShowPage(data))
	if err != nil {
		slog.Error("expense.openPaidAtDialog: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, data.Signals)
	return c.NoContent(http.StatusOK)
}

func openPaidAtDialogInIndex(c echo.Context, groupID, id string, tableQuery utils.TableQuery) error {
	query := utils.NormalizeTableQuery(tableQuery, TableQuerySpec())
	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.openPaidAtDialogInIndex: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	paidAtValue := ""
	expenseTitle := ""
	found := false
	for _, expense := range data.Expenses {
		if expense.ID != id {
			continue
		}
		found = true
		expenseTitle = strings.TrimSpace(expense.Title)
		if expense.PaidAt.Valid {
			paidAtValue = utils.FormatDateInput(expense.PaidAt.String)
		}
		break
	}
	if !found {
		slog.Info("expense.openPaidAtDialogInIndex: expense not found in page data", "expense_id", id)
	}

	data.PaidAtDialog = PaidAtDialogState{
		Open:        true,
		Mode:        "table",
		Title:       ctxi18n.T(c.Request().Context(), "fields.paid_at"),
		Message:     expenseTitle,
		Value:       paidAtValue,
		SubmitLabel: ctxi18n.T(c.Request().Context(), "table.apply"),
		CancelLabel: ctxi18n.T(c.Request().Context(), "actions.cancel"),
		URL:         "/groups/" + groupID + "/expenses/" + id + "/paid_at",
		TriggerID:   "expense-index-paid-at-edit",
	}
	data.Signals = expenseIndexSignals(data.Query)
	if paidAtDialog, ok := data.Signals["paidAtDialog"].(map[string]any); ok {
		paidAtDialog["open"] = data.PaidAtDialog.Open
		paidAtDialog["fetching"] = data.PaidAtDialog.Fetching
		paidAtDialog["mode"] = data.PaidAtDialog.Mode
		paidAtDialog["title"] = data.PaidAtDialog.Title
		paidAtDialog["message"] = data.PaidAtDialog.Message
		paidAtDialog["value"] = data.PaidAtDialog.Value
		paidAtDialog["submitLabel"] = data.PaidAtDialog.SubmitLabel
		paidAtDialog["cancelLabel"] = data.PaidAtDialog.CancelLabel
		paidAtDialog["url"] = data.PaidAtDialog.URL
		paidAtDialog["triggerID"] = data.PaidAtDialog.TriggerID
	}

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndexPage(data))
	if err != nil {
		slog.Error("expense.openPaidAtDialogInIndex: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	utils.SSEHub.PatchSignals(c, data.Signals)
	return c.NoContent(http.StatusOK)
}

func patchUpdatePaidAt(c echo.Context, groupID, id, mode string, tableQuery utils.TableQuery, value string) error {
	paidAt := normalizePaidAtInput(value)

	expense, err := db.GetExpense(c.Request().Context(), db.GetExpenseParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.updatePaidAt: failed to get expense", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	paid := expense.Paid
	if paid == 0 && paidAt != "" {
		paid = 1
	}

	_, err = db.UpdateExpense(c.Request().Context(), db.UpdateExpenseParams{
		Title:       expense.Title,
		Description: expense.Description,
		Amount:      expense.Amount,
		Date:        expense.Date,
		Paid:        paid,
		PaidAt:      paidAt,
		ID:          id,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("expense.updatePaidAt: failed to update paid_at", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))
	utils.InvalidateGroupCaches(groupID)

	if mode == "table" {
		query := utils.NormalizeTableQuery(tableQuery, TableQuerySpec())
		data, err := GetIndexData(c.Request().Context(), groupID, query)
		if err != nil {
			slog.Error("expense.updatePaidAt: failed to get index data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		applyExpenseTableByRole(&data, middleware.IsAdmin(c))
		data.Signals = expenseIndexSignals(data.Query)
		data.IsAuthenticated = true
		data.IsSuperAdmin = middleware.IsSuperadmin(c)
		html, err := utils.RenderHTMLForRequest(c, ExpenseIndexPage(data))
		if err != nil {
			slog.Error("expense.updatePaidAt: failed to render index", "err", err)
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

	data, err := GetShowData(c.Request().Context(), groupID, id)
	if err != nil {
		slog.Error("expense.updatePaidAt: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.Signals = expenseShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, ExpenseShowPage(data))
	if err != nil {
		slog.Error("expense.updatePaidAt: failed to render", "err", err)
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

func OpenPaidAtPrompt(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.openPaidAtPrompt: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	query := utils.ParseTableQuery(c, staticTableQueryable{spec: TableQuerySpec()})
	mode := "single"
	var signals modeParams
	if err := datastar.ReadSignals(c.Request(), &signals); err == nil {
		if utils.SetTabID(c, signals.TabID) {
			query = utils.NormalizeTableQuery(signals.TableQuery, TableQuerySpec())
			mode = strings.TrimSpace(signals.Mode)
		}
	}
	if mode == "table" {
		return openPaidAtDialogInIndex(c, groupID, id, query)
	}

	return openPaidAtDialog(c, groupID, id, query)
}

func UpdatePaidAt(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.updatePaidAt: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	var signals paidAtParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.updatePaidAt: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	return patchUpdatePaidAt(c, groupID, id, signals.Mode, signals.TableQuery, signals.PaidAtDialog.Value)
}

func TogglePaid(c echo.Context) error {
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

	_, err = db.ToggleExpensePaid(c.Request().Context(), db.ToggleExpensePaidParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.togglePaid: failed to toggle paid status", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "expenses.notifications.toggle_paid_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	slog.Debug("expense.togglePaid", "id", id)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	if signals.Mode == "single" {
		data, err := GetShowData(c.Request().Context(), groupID, id)
		if err != nil {
			slog.Error("expense.togglePaid: failed to get data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		data.IsAdmin = middleware.IsAdmin(c)
		data.Signals = expenseShowSignals(data)
		data.IsAuthenticated = true
		data.IsSuperAdmin = middleware.IsSuperadmin(c)
		html, err := utils.RenderHTMLForRequest(c, ExpenseShowPage(data))
		if err != nil {
			slog.Error("expense.togglePaid: failed to render", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	query := utils.NormalizeTableQuery(signals.TableQuery, TableQuerySpec())
	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.togglePaid: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)
	html, err := utils.RenderHTMLForRequest(c, ExpenseIndexPage(data))
	if err != nil {
		slog.Error("expense.togglePaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}
