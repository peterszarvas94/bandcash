package settings

import (
	"net/http"

	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

func (s *Settings) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	return utils.RenderComponent(c, SettingsIndex(s.Data(c.Request().Context())))
}

func (s *Settings) UpdateLanguage(c echo.Context) error {
	locale := appi18n.NormalizeLocale(c.FormValue(appi18n.CookieName))
	c.SetCookie(&http.Cookie{
		Name:     appi18n.CookieName,
		Value:    locale,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	return c.Redirect(http.StatusSeeOther, "/settings")
}
