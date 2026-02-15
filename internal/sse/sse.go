package sse

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/utils"
)

func Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		log := slog.Default()

		clientID, err := utils.GetClientID(c)
		if err != nil {
			log.Warn("sse: no client_id cookie")
			return c.String(400, "No client_id cookie")
		}

		sseConn := datastar.NewSSE(w, r)
		SSEHub.AddClient(clientID, sseConn)

		log.Debug("sse: client connected", "client_id", clientID)

		defer func() {
			SSEHub.RemoveClient(clientID)
			log.Debug("sse: client disconnected", "client_id", clientID)
		}()

		// Keep connection alive until client disconnects
		<-r.Context().Done()
		return nil
	}
}
