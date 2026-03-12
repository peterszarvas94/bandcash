package i18n

import (
	"context"
	"embed"
	"net/http"
	"net/url"
	"strings"

	"github.com/invopop/ctxi18n"
	ctxi18nCore "github.com/invopop/ctxi18n/i18n"
)

const (
	DefaultLocale = "hu"
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

func LocaleFromRequest(r *http.Request) string {
	if r == nil {
		return DefaultLocale
	}

	if lang := NormalizeLocale(r.URL.Query().Get("lang")); lang != "" {
		if r.URL.Query().Get("lang") != "" {
			return lang
		}
	}

	acceptLanguage := strings.TrimSpace(r.Header.Get("Accept-Language"))
	if acceptLanguage == "" {
		return DefaultLocale
	}

	for _, part := range strings.Split(acceptLanguage, ",") {
		candidate := strings.TrimSpace(strings.Split(part, ";")[0])
		if candidate == "" {
			continue
		}
		return NormalizeLocale(candidate)
	}

	return DefaultLocale
}

func LocalizedHomePath(ctx context.Context) string {
	return "/?lang=" + url.QueryEscape(LocaleCode(ctx))
}
