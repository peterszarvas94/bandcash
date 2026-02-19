package settings

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	appi18n "bandcash/internal/i18n"
	"bandcash/internal/utils"
)

type settingsSignals struct {
	FormData struct {
		Lang string `json:"lang"`
	} `json:"formData"`
}

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
	signals := settingsSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	locale := appi18n.NormalizeLocale(signals.FormData.Lang)
	c.SetCookie(&http.Cookie{
		Name:     appi18n.CookieName,
		Value:    locale,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 365,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	})
	err := utils.SSEHub.Redirect(c, "/settings")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}
