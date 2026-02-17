package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

type Home struct {
}

// Data returns data for rendering.
func (h *Home) Data() HomeData {
	return HomeData{
		Title:       "Bandcash",
		Breadcrumbs: []utils.Crumb{},
	}
}

// Register registers home routes.
func Register(e *echo.Echo) {
	h := &Home{}
	e.GET("/", h.Index)
}
