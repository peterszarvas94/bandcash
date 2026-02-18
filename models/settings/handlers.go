package settings

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

func (s *Settings) Index(c echo.Context) error {
	utils.EnsureClientID(c)
	data := s.Data(c.Request().Context())
	if cookie, err := c.Cookie("session"); err == nil && cookie.Value != "" {
		if user, err := db.Qry.GetUserByID(c.Request().Context(), cookie.Value); err == nil {
			data.UserEmail = user.Email
		}
	}
	return utils.RenderComponent(c, SettingsIndex(data))
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
