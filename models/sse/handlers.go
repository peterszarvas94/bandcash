package sse

import (
	"bandcash/internal/utils"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

func SSEHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		log := slog.Default()

		clientID, err := utils.GetClientID(c)
		if err != nil {
			log.Warn("sse: no client_id cookie")
			return c.NoContent(400)
		}

		sseConn := datastar.NewSSE(w, r)
		utils.SSEHub.AddClient(clientID, sseConn)

		log.Debug("sse: client connected", "client_id", clientID)

		defer func() {
			utils.SSEHub.RemoveClient(clientID)
			log.Debug("sse: client disconnected", "client_id", clientID)
		}()

		// Keep connection alive until client disconnects
		<-r.Context().Done()
		return nil
	}
}
