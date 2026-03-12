package middleware

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func RequestLogger() echo.MiddlewareFunc {
	return echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    false,
		LogValuesFunc: func(c echo.Context, v echoMiddleware.RequestLoggerValues) error {
			req := c.Request()
			slog.Info("http.request.completed",
				"path", req.URL.Path,
				"query", req.URL.RawQuery,
				"method", req.Method,
				"status", v.Status,
			)
			return nil
		},
	})
}
