package health

import "github.com/labstack/echo/v4"

// Register registers health routes.
func Register(e *echo.Echo) {
	h := &Health{}
	e.GET("/health", h.Check)
}
