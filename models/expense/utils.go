package expense

import (
	"fmt"
	"net/url"
	"strings"
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
