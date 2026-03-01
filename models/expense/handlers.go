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
		"mode":      "",
		"formState": "",
		"editingId": "",
		"formData":  map[string]any{"title": "", "description": "", "amount": 0, "date": ""},
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
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	return utils.RenderPage(c, ExpenseIndex(data))
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
	})
	if err != nil {
		slog.Error("expense.create.table: failed to create expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.create_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.created"))
	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.create.table: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
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
		ID:          id,
		GroupID:     groupID,
	})
	if err != nil {
		slog.Error("expense.update: failed to update expense", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "expenses.notifications.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "expenses.notifications.updated"))
	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.update: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
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
	utils.SSEHub.PatchSignals(c, defaultExpenseSignals)

	query := utils.NormalizeTableQuery(signals.TableQuery, e.TableQuerySpec())
	data, err := e.GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("expense.destroy: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.UserEmail = userEmail

	html, err := utils.RenderHTMLForRequest(c, ExpenseIndex(data))
	if err != nil {
		slog.Error("expense.destroy: failed to render", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}
