package middleware

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/billing"
	"bandcash/internal/utils"
)

func RequireCanCreateGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if utils.IsSuperadmin(c) {
			return next(c)
		}
		allowed, _, err := billing.CanOwnAnotherGroup(c.Request().Context(), utils.GetUserID(c))
		if err != nil {
			slog.Error("billing.middleware.group: check failed", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if !allowed {
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "billing.errors.subscription_slots_exhausted"))
			return c.Redirect(http.StatusFound, "/pricing")
		}
		return next(c)
	}
}

func RequireCanCreateEvent(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}
