package sse

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

func Register(e *echo.Echo) {
	e.GET("/sse", utils.SSEHandler())
}
