package sse

import (
	"bandcash/internal/utils"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func SSEHandler() echo.HandlerFunc {
	type sseSignals struct {
		TabID string `json:"tab_id"`
	}

	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		log := slog.Default()

		signals := sseSignals{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			log.Warn("sse: failed to read signals", "err", err)
			return c.NoContent(http.StatusBadRequest)
		}

		if !utils.SetTabID(c, signals.TabID) {
			log.Warn("sse: invalid tab_id in signals")
			return c.NoContent(http.StatusBadRequest)
		}

		tabIDValue := utils.TabIDFromContext(c.Request().Context())

		sseConn := datastar.NewSSE(w, r)
		utils.SSEHub.AddClient(tabIDValue, sseConn)

		log.Debug("sse: client connected", "tab_id", tabIDValue)

		defer func() {
			utils.SSEHub.RemoveClient(tabIDValue)
			log.Debug("sse: client disconnected", "tab_id", tabIDValue)
		}()

		// Keep connection alive until client disconnects
		<-r.Context().Done()
		return nil
	}
}
