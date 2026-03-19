package utils

import (
	"context"

	appi18n "bandcash/internal/i18n"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatNumberLocalized(ctx context.Context, value int64) string {
	printer := numberPrinterByLocale(ctx)
	return printer.Sprintf("%d", value)
}

func FormatNumberLocalizedWithSign(ctx context.Context, value int64, positive bool) string {
	sign := "-"
	if positive {
		sign = "+"
	}
	return sign + " " + FormatNumberLocalized(ctx, value)
}

func numberPrinterByLocale(ctx context.Context) *message.Printer {
	switch appi18n.LocaleCode(ctx) {
	case "hu":
		return message.NewPrinter(language.Hungarian)
	default:
		return message.NewPrinter(language.English)
	}
}
