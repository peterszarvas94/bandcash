package health

import "github.com/labstack/echo/v4"

// RegisterRoutes registers health routes.
func RegisterRoutes(e *echo.Echo) {
	h := &Health{}
	e.GET("/health", h.Check)
}
