package dev

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
)

func RegisterRoutes(e *echo.Echo) {
	if utils.Env().AppEnv != "development" {
		return
	}

	h := &DevNotifications{}
	g := e.Group("/dev")
	g.GET("", h.DevPageHandler)
	g.POST("/spinner", h.TestSpinner)
	g.POST("/multi-action/:action", h.TestMultiAction)
	g.POST("/notifications/inline", h.TestInline)
	g.POST("/notifications/test", h.Test)
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
	g.GET("/errors/500", h.PreviewInternalErrorPage)
}
