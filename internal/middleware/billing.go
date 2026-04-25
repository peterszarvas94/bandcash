package middleware

import (
	"log/slog"
	"net/http"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/flags"
	"bandcash/internal/utils"
	"github.com/labstack/echo/v4"
)

func RequireCanCreateGroup(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func RequireCanCreateEvent(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		return next(c)
	}
}

func RequireWithinSubscriptionLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if utils.IsSuperadmin(c) {
			allowBypass, err := flags.IsBypassLimitForSuperadminEnabled(c.Request().Context())
			if err != nil {
				slog.Error("billing gate: failed to read superadmin bypass flag", "err", err)
				return c.Redirect(http.StatusFound, "/over-limit")
			}
			if allowBypass {
				return next(c)
			}
		}

		userID := utils.GetUserID(c)
		state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID)
		if err != nil {
			slog.Error("billing gate: failed to load access state", "user_id", userID, "err", err)
			return c.Redirect(http.StatusFound, "/over-limit")
		}
		if internalbilling.IsLimitExceeded(state) {
			return c.Redirect(http.StatusFound, "/over-limit")
		}
		return next(c)
	}
}
