package static

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo) {
	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")
}
