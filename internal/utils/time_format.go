package utils

import (
	"context"
	"time"

	appi18n "bandcash/internal/i18n"
)

func FormatDateTimeLocalized(ctx context.Context, value string) string {
	t, ok := parseDateTime(value)
	if !ok {
		return value
	}
	return formatDateTimeByLocale(ctx, t)
}

func FormatTimeLocalized(ctx context.Context, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return formatDateTimeByLocale(ctx, t)
}

func formatDateTimeByLocale(ctx context.Context, t time.Time) string {
	switch appi18n.LocaleCode(ctx) {
	case "hu":
		return t.Format("2006. 01. 02. 15:04")
	default:
		return t.Format("Jan 2, 2006 3:04 PM")
	}
}

func parseDateTime(value string) (time.Time, bool) {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	} {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
