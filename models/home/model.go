package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	homeview "bandcash/models/home/templates/view"
)

type Home struct {
}

// Data returns data for rendering.
func (h *Home) Data() homeview.HomeData {
	return homeview.HomeData{
		Title:       "Bandcash",
		Breadcrumbs: []utils.Crumb{},
	}
}

// Register registers home routes.
func Register(e *echo.Echo) {
	h := &Home{}
	e.GET("/", h.Index)
}
