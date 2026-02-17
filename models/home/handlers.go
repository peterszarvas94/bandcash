package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	homepages "bandcash/models/home/templates/pages"
)

// Index renders the home page with links to examples.
func (h *Home) Index(c echo.Context) error {
	return utils.RenderComponent(c, homepages.HomeIndex(h.Data()))
}
