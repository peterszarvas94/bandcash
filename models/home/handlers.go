package home

import (
	"net/url"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

// Index renders the home page with welcome message.
func (h *Home) Index(c echo.Context) error {
	utils.EnsureClientID(c)

	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if _, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			if c.QueryParam("lang") != "" {
				return c.Redirect(302, "/dashboard?lang="+url.QueryEscape(appi18n.NormalizeLocale(c.QueryParam("lang"))))
			}
			return c.Redirect(302, "/dashboard")
		}
	}

	data := h.Data(c.Request().Context())
	return utils.RenderPage(c, HomeIndex(data))
}
