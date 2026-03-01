package utils

import (
	"bytes"
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func renderContext(c echo.Context) context.Context {
	ctx := c.Request().Context()
	clientID := EnsureClientID(c)
	includeActive := c.Request().Method != echo.GET
	items := Notifications.DrainForRender(clientID, includeActive)
	return WithNotifications(ctx, items)
}

// RenderPage writes a templ component to the response.
func RenderPage(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(renderContext(c), c.Response().Writer)
}

// RenderHTML renders a templ component to a string.
func RenderHTML(ctx context.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderHTMLForRequest(c echo.Context, component templ.Component) (string, error) {
	return RenderHTML(renderContext(c), component)
}
