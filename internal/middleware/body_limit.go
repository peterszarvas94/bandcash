package middleware

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/labstack/echo/v4"
)

func GlobalBodyLimit() echo.MiddlewareFunc {
	return echoMiddleware.BodyLimit("1M")
}

func AuthBodyLimit() echo.MiddlewareFunc {
	return echoMiddleware.BodyLimit("64K")
}
