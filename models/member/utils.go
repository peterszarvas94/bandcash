package member

import (
	"strings"

	"bandcash/internal/utils"
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

func applyMemberShowTableByRole(data *MemberData, isAdmin bool) {
	data.IsAdmin = isAdmin
	if !isAdmin {
		data.EventsTable.ActionsWidthRem = 0
	}
}
