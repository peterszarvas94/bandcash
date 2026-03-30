package utils

import (
	"net/http"

	"github.com/labstack/echo/v4"

	appi18n "bandcash/internal/i18n"
)

const (
	SessionCookieName = "session"
	CSRFCookieName    = "_csrf"
)

func SetSessionCookie(c echo.Context, token string) {
	env := Env()
	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   env.AppEnv == "production" || env.AppEnv == "staging",
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func SetCSRFCookie(c echo.Context, token string) {
	env := Env()
	c.SetCookie(&http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   env.AppEnv == "production" || env.AppEnv == "staging",
		SameSite: http.SameSiteLaxMode,
	})
}

func SetLocaleCookie(c echo.Context, locale string) {
	env := Env()
	c.SetCookie(&http.Cookie{
		Name:     appi18n.LocaleCookieName,
		Value:    appi18n.NormalizeLocale(locale),
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: true,
		Secure:   env.AppEnv == "production" || env.AppEnv == "staging",
		SameSite: http.SameSiteLaxMode,
	})
}
