package event

import (
	"fmt"
	"net/url"
	"strings"

	"bandcash/internal/db"
	"bandcash/internal/utils"
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

func eventDateValue(event db.Event) string {
	date := strings.TrimSpace(event.Date)
	if date != "" {
		return date
	}
	return utils.FormatDateInput(event.Time)
}

func eventTimeValue(event db.Event) string {
	eventTime := strings.TrimSpace(event.EventTime)
	if eventTime != "" {
		return eventTime
	}
	trimmed := strings.TrimSpace(event.Time)
	if len(trimmed) >= 16 {
		return trimmed[11:16]
	}
	return ""
}

func eventDateTimeValue(event db.Event) string {
	date := eventDateValue(event)
	eventTime := eventTimeValue(event)
	if date == "" {
		return strings.TrimSpace(event.Time)
	}
	if eventTime == "" {
		return date
	}
	return date + "T" + eventTime
}
