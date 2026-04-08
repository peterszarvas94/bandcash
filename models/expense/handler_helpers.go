package expense

import (
	"database/sql"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
}

func getUserEmail(c echo.Context) string {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return ""
	}
	user, err := db.GetUserByID(c.Request().Context(), userID)
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
