package sse

import (
	"bytes"
	"html/template"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	appmw "webapp/internal/middleware"
	"webapp/internal/utils"
)

// DataProvider is a function that returns data to render in the SSE response.
type DataProvider func() any

// Handler creates an SSE handler that uses the provided template, template name and data function.
func Handler(tmpl *template.Template, templateName string, getData DataProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response().Writer
		log := appmw.Logger(c)

		clientID, err := utils.GetClientID(c)
		if err != nil {
			log.Warn("sse: no client_id cookie")
			return c.String(400, "No client_id cookie")
		}

		log.Debug("sse: client connected")

		sse := datastar.NewSSE(w, r)
		client := utils.Store.AddClient(clientID, sse)

		// Render template to string
		render := func() (string, error) {
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, templateName, getData()); err != nil {
				return "", err
			}
			return buf.String(), nil
		}

		// Send initial body
		html, err := render()
		if err != nil {
			log.Error("sse: template error", "err", err)
			return c.NoContent(500)
		}
		if err := sse.PatchElements(html); err != nil {
			log.Error("sse: initial patch error", "err", err)
			return c.NoContent(500)
		}
		log.Debug("sse: initial app sent")

		// Cleanup on disconnect
		defer func() {
			utils.Store.RemoveClient(clientID)
			log.Debug("sse: client disconnected")
		}()

		// Wait for signals and send updates
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

				html, err := render()
				if err != nil {
					log.Error("sse: template error", "err", err)
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
