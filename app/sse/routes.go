package sse

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/sse"
)

func Register(e *echo.Echo) {
	e.GET("/sse", sse.Handler())
}
