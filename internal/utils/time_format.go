package utils

import (
	"context"
	"strings"
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

func FormatDateLocalized(ctx context.Context, value string) string {
	t, ok := parseDate(value)
	if !ok {
		return value
	}
	return formatDateByLocale(ctx, t)
}

func FormatDateTimeInput(value string) string {
	t, ok := parseDateTime(value)
	if !ok {
		return ""
	}
	return t.Format("2006-01-02T15:04")
}

func FormatDateInput(value string) string {
	if t, ok := parseDateTime(value); ok {
		return t.Format("2006-01-02")
	}

	if t, ok := parseDate(value); ok {
		return t.Format("2006-01-02")
	}

	return ""
}

func FormatTimeLocalized(ctx context.Context, t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return formatDateTimeByLocale(ctx, t)
}

func FormatClockLocalized(ctx context.Context, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	for _, layout := range []string{"15:04", "15:04:05"} {
		t, err := time.Parse(layout, trimmed)
		if err != nil {
			continue
		}
		switch appi18n.LocaleCode(ctx) {
		case "hu":
			return t.Format("15:04")
		default:
			return t.Format("3:04 PM")
		}
	}
	return trimmed
}

func formatDateTimeByLocale(ctx context.Context, t time.Time) string {
	switch appi18n.LocaleCode(ctx) {
	case "hu":
		return t.Format("2006. 01. 02. 15:04")
	default:
		return t.Format("Jan 2, 2006 3:04 PM")
	}
}

func formatDateByLocale(ctx context.Context, t time.Time) string {
	switch appi18n.LocaleCode(ctx) {
	case "hu":
		return t.Format("2006. 01. 02.")
	default:
		return t.Format("Jan 2, 2006")
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

func parseDate(value string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
