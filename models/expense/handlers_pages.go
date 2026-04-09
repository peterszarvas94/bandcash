package expense

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	expensestore "bandcash/models/expense/store"
	groupstore "bandcash/models/group/store"
)

func NewExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
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
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}
	return utils.RenderPage(c, ExpenseNewPage(data))
}

func EditExpensePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixExpense) {
		slog.Info("expense.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("expense.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	expense, err := expensestore.GetExpense(c.Request().Context(), expensestore.GetExpenseParams{
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
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}
	return utils.RenderPage(c, ExpenseEditPage(data))
}

func IndexPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: TableQuerySpec()})

	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, utils.IsAdmin(c))
	data.Signals = expenseIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseIndexPage(data))
}

func ShowPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

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
	data.IsAdmin = utils.IsAdmin(c)
	data.Signals = expenseShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, ExpenseShowPage(data))
}
