package middleware

import (
	"log/slog"

	"github.com/invopop/ctxi18n"
	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
)

func Locale() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			locale := appi18n.DefaultLocale
			cookie, err := c.Cookie(appi18n.CookieName)
			if err == nil && cookie.Value != "" {
				locale = cookie.Value
			}
			locale = appi18n.NormalizeLocale(locale)

			ctx, err := ctxi18n.WithLocale(c.Request().Context(), locale)
			if err != nil {
				slog.Error("i18n.locale: failed to set locale", "err", err)
				return c.NoContent(500)
			}
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
