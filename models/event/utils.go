package event

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

func eventsCachePrefix(groupID string) string {
	return fmt.Sprintf("events_group_%s_", normalizeCacheKeyPart(groupID))
}

func EventsFilterKey(groupID, search, year, from, to, sort, dir string) string {
	return fmt.Sprintf("%ssearch_%s_year_%s_from_%s_to_%s_sort_%s_dir_%s",
		eventsCachePrefix(groupID),
		normalizeCacheKeyPart(search),
		normalizeCacheKeyPart(year),
		normalizeCacheKeyPart(from),
		normalizeCacheKeyPart(to),
		normalizeCacheKeyPart(sort),
		normalizeCacheKeyPart(dir),
	)
}
