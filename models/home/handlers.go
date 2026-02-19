package home

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

// Index renders the home page with welcome message.
func (h *Home) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	data := h.Data(c.Request().Context())

	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if _, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			data.UserLoggedIn = true
		}
	}

	return utils.RenderComponent(c, HomeIndex(data))
}
