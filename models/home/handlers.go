package home

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

// Index renders the home page with links to examples.
func (h *Home) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	userEmail := ""

	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if _, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			return c.Redirect(http.StatusFound, "/groups")
		}
	}
	data := h.Data(c.Request().Context())
	data.UserEmail = userEmail
	return utils.RenderComponent(c, HomeIndex(data))
}
