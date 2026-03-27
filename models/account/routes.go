package account

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Account {
	account := New()

	e.GET("/language", account.LanguagePage, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/account", account.Index, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/account/language", account.UpdateLanguage, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/account/details-state", account.UpdateDetailsState, middleware.RequireAuth, middleware.WithDetailState)
	e.DELETE("/account/sessions/:id", account.LogoutSession, middleware.RequireAuth, middleware.WithDetailState)
	e.DELETE("/account/sessions", account.LogoutAllOtherSessions, middleware.RequireAuth, middleware.WithDetailState)

	return account
}
