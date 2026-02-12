package sse

import (
	"encoding/json"
	"log/slog"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/hub"
	"bandcash/internal/utils"
)

type ViewRenderer func(c echo.Context, view string) (string, error)

type viewSignals struct {
	View json.RawMessage `json:"view"`
}

func HandlerWithView(render ViewRenderer) echo.HandlerFunc {
	// HandlerWithView closes over the render function so one SSE handler can
	// render different views based on the per-client view key.
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		log := slog.Default()

		clientID, err := utils.GetClientID(c)
		if err != nil {
			log.Warn("sse: no client_id cookie")
			return c.String(400, "No client_id cookie")
		}

		var signals viewSignals
		if err := datastar.ReadSignals(r, &signals); err != nil {
			log.Warn("sse: failed to read signals", "err", err)
			return c.NoContent(400)
		}

		view, err := utils.ParseRawString(signals.View)
		if err != nil || view == "" {
			log.Warn("sse: missing view", "err", err)
			return c.NoContent(400)
		}

		// Persist the view key for this client so updates can re-render
		// the current route without needing the request path again.
		hub.Hub.SetView(clientID, view)

		log.Debug("sse: client connected", "view", view)

		sse := datastar.NewSSE(w, r)
		client := hub.Hub.AddClient(clientID, sse)

		renderView := func() (string, error) {
			// Resolve the current view for this client and render it into HTML.
			currentView, ok := hub.Hub.GetView(clientID)
			if !ok {
				return "", echo.NewHTTPError(404, "view not set")
			}
			return render(c, currentView)
		}

		html, err := renderView()
		if err != nil {
			log.Error("sse: render error", "err", err)
			return c.NoContent(500)
		}
		if err := sse.PatchElements(html); err != nil {
			log.Error("sse: initial patch error", "err", err)
			return c.NoContent(500)
		}
		log.Debug("sse: initial app sent")

		// Cleanup on disconnect.
		defer func() {
			hub.Hub.RemoveClient(clientID)
			log.Debug("sse: client disconnected")
		}()

		// Wait for signals and send updates. Updates are triggered by handlers
		// calling hub.Hub.Render(c) for the current client.
		for {
			select {
			case <-r.Context().Done():
				log.Debug("sse: context done")
				return nil
			case _, ok := <-client.Signals:
				if !ok {
					log.Debug("sse: signal channel closed")
					return nil
				}

				html, err := renderView()
				if err != nil {
					log.Error("sse: render error", "err", err)
					continue
				}

				if err := sse.PatchElements(html); err != nil {
					log.Error("sse: patch error", "err", err)
					return nil
				}
				log.Debug("sse: update sent")
			}
		}
	}
}
