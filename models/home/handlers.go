package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

// Index renders the home page with links to examples.
func (h *Home) Index(c echo.Context) error {
	return utils.RenderComponent(c, HomeIndex(h.Data()))
}
