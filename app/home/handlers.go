package home

import "github.com/labstack/echo/v4"

// Index renders the home page with links to examples.
func (h *Home) Index(c echo.Context) error {
	return h.tmpl.ExecuteTemplate(c.Response().Writer, "index", h.Data())
}
