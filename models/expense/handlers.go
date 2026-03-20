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

type expenseParams struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Date        string `json:"date" validate:"required"`
	Paid        bool   `json:"paid"`
}

type expenseTableParams struct {
	FormData   expenseParams    `json:"formData"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	Mode       string           `json:"mode"`
}

type modeParams struct {
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

var (
	defaultExpenseSignals = map[string]any{
		"mode":      "table",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false},
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

func (e *Expenses) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)
	query := utils.ParseTableQuery(c, e)

	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	return utils.RenderPage(c, ExpenseIndex(data))
}

func (e *Expenses) Show(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

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
	data.UserEmail = userEmail

	return utils.RenderPage(c, ExpenseShow(data))
}

func (e *Expenses) Create(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	var signals expenseTableParams
	err := datastar.ReadSignals(c.Request(), &signals)
	if err != nil {
		slog.Info("expense.create.table: failed to read signals", "err", err)
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.Date = strings.TrimSpace(signals.FormData.Date)

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
	})
	if err != nil {
		slog.Error("expense.create.table: failed to create expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.created"))
	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(utils.TableQuery{}, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.create.table: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.create.table: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Expenses) Update(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

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
	signals.FormData.Title = strings.TrimSpace(signals.FormData.Title)
	signals.FormData.Description = strings.TrimSpace(signals.FormData.Description)
	signals.FormData.Date = strings.TrimSpace(signals.FormData.Date)

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
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("expense.update: failed to update expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))

	if signals.Mode == "single" {
		utils.SSEHub.PatchSignals(c, map[string]any{
			"formState": "",
			"formData": map[string]any{
				"title":       signals.FormData.Title,
				"description": signals.FormData.Description,
				"amount":      signals.FormData.Amount,
				"date":        signals.FormData.Date,
				"paid":        signals.FormData.Paid,
			},
			"errors": map[string]any{"title": "", "description": "", "amount": "", "date": ""},
		})
		data, err := e.GetShowData(c.Request().Context(), groupID, id)
		if err != nil {
			slog.Error("expense.update: failed to get data", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		data.IsAdmin = middleware.IsAdmin(c)
		data.UserEmail = userEmail
		html, err := utils.RenderHTMLForRequest(c, ExpenseShow(data))
		if err != nil {
			slog.Error("expense.update: failed to render", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	// Clear cache to ensure fresh data on next load
	utils.InvalidateGroupCaches(groupID)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.update: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyExpenseTableByRole(&data, middleware.IsAdmin(c))
	data.UserEmail = userEmail

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.update: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}

func (e *Expenses) Destroy(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

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
	data.UserEmail = userEmail

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
	userEmail := getUserEmail(c)

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
		data.UserEmail = userEmail
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
	data.UserEmail = userEmail
	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.togglePaid: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}
