package member

import (
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	authstore "bandcash/models/auth/store"
)

type staticTableQueryable struct {
	spec utils.TableQuerySpec
}

func (s staticTableQueryable) TableQuerySpec() utils.TableQuerySpec {
	return s.spec
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

func applyMemberShowTableByRole(data *MemberData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.EventsTable.ActionsWidthRem = 0
	}
}
