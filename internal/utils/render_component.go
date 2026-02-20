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
	items := Notifications.Drain(clientID)
	return WithNotifications(ctx, items)
}

// RenderComponent writes a templ component to the response.
func RenderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(renderContext(c), c.Response().Writer)
}

// RenderComponentString renders a templ component to a string.
func RenderComponentString(ctx context.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderComponentStringFor(c echo.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(renderContext(c), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
