package expense

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
)

func normalizeCacheKeyPart(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "all"
	}
	return url.QueryEscape(trimmed)
}

func expensesCachePrefix(groupID string) string {
	return fmt.Sprintf("expenses_group_%s_", normalizeCacheKeyPart(groupID))
}

func ExpensesFilterKey(groupID, search, year, from, to, sort, dir string) string {
	return fmt.Sprintf("%ssearch_%s_year_%s_from_%s_to_%s_sort_%s_dir_%s",
		expensesCachePrefix(groupID),
		normalizeCacheKeyPart(search),
		normalizeCacheKeyPart(year),
		normalizeCacheKeyPart(from),
		normalizeCacheKeyPart(to),
		normalizeCacheKeyPart(sort),
		normalizeCacheKeyPart(dir),
	)
}

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
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
