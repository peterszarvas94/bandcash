package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

func WithDetailState(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID := GetUserID(c)
		if userID != "" {
			loadDetailCardStates(c, userID)
		}
		return next(c)
	}
}

func loadDetailCardStates(c echo.Context, userID string) {
	states, err := db.ListUserDetailCardStates(c.Request().Context(), userID)
	if err != nil {
		slog.Warn("middleware.detail_card_states: failed to load", "user_id", userID, "err", err)
		return
	}

	statesMap := make(map[string]bool, len(states))
	for _, state := range states {
		statesMap[state.StateKey] = state.IsOpen == 1
	}

	c.SetRequest(c.Request().WithContext(utils.WithDetailCardStates(c.Request().Context(), statesMap)))
}
