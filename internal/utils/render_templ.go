package utils

import (
	"bytes"
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// RenderComponent writes a templ component to the response.
func RenderComponent(c echo.Context, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// RenderComponentString renders a templ component to a string.
func RenderComponentString(ctx context.Context, component templ.Component) (string, error) {
	var buf bytes.Buffer
	if err := component.Render(ctx, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
