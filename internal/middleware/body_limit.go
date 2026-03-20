package middleware

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func GlobalBodyLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return echoMiddleware.BodyLimit("1M")(next)
}

func AuthBodyLimit(next echo.HandlerFunc) echo.HandlerFunc {
	return echoMiddleware.BodyLimit("64K")(next)
}
