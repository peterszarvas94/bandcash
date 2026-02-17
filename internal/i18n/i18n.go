package i18n

import (
	"context"
	"embed"

	"github.com/invopop/ctxi18n"
	ctxi18nCore "github.com/invopop/ctxi18n/i18n"
)

const (
	DefaultLocale = "en"
	CookieName    = "lang"
)

var SupportedLocales = []string{"en", "hu"}

//go:embed locales/**
var localesFS embed.FS

func Load() error {
	return ctxi18n.LoadWithDefault(localesFS, ctxi18nCore.Code(DefaultLocale))
}

func LocaleCode(ctx context.Context) string {
	l := ctxi18n.Locale(ctx)
	if l == nil {
		return DefaultLocale
	}
	return string(l.Code())
}

func NormalizeLocale(code string) string {
	if code == "" {
		return DefaultLocale
	}
	l := ctxi18n.Match(code)
	if l == nil {
		return DefaultLocale
	}
	return string(l.Code())
}
