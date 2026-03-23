package dev

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

func RegisterRoutes(e *echo.Echo) {
	if utils.Env().AppEnv != "development" {
		return
	}

	h := &DevNotifications{}
	g := e.Group("/dev", middleware.RequireAuth)
	g.GET("", h.DevPageHandler)
	g.POST("/body-limit/global", h.TestBodyLimitGlobal)
	g.POST("/body-limit/auth", h.TestBodyLimitAuth, middleware.AuthBodyLimit)
	g.POST("/spinner", h.TestSpinner)
	g.POST("/multi-action/:action", h.TestMultiAction)
	g.POST("/notifications/inline", h.TestInline)
	g.POST("/notifications/success", h.TestSuccess)
	g.POST("/notifications/error", h.TestError)
	g.POST("/notifications/info", h.TestInfo)
	g.GET("/emails/login", h.PreviewLoginEmail)
	g.GET("/emails/invite", h.PreviewInviteEmail)
	g.GET("/emails/invite-accepted", h.PreviewInviteAcceptedEmail)
	g.GET("/emails/group-created", h.PreviewGroupCreatedEmail)
	g.GET("/emails/role-upgraded", h.PreviewRoleUpgradedEmail)
	g.GET("/emails/role-downgraded", h.PreviewRoleDowngradedEmail)
	g.GET("/emails/access-removed", h.PreviewAccessRemovedEmail)
	g.GET("/errors/link-invalid", h.PreviewInvalidLinkErrorPage)
	g.GET("/errors/400", h.PreviewBadRequestErrorPage)
	g.GET("/errors/403", h.PreviewForbiddenErrorPage)
	g.GET("/errors/404", h.PreviewNotFoundErrorPage)
	g.GET("/errors/429", h.PreviewRateLimitErrorPage)
	g.GET("/errors/500", h.PreviewInternalErrorPage)
	g.GET("/query-test/:model", h.TestTableQuery)
}
