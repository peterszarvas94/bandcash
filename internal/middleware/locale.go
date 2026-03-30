package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

func Locale(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		locale := appi18n.LocaleFromRequest(c.Request())
		if rawLang := strings.TrimSpace(c.QueryParam("lang")); rawLang != "" {
			utils.SetLocaleCookie(c, rawLang)
		}

		ctx, err := ctxi18n.WithLocale(c.Request().Context(), locale)
		if err != nil {
			slog.Error("i18n.locale: failed to set locale", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}
