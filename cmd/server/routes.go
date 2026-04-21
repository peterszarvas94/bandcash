package main

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	"bandcash/models/account"
	"bandcash/models/admin"
	"bandcash/models/auth"
	billingmodel "bandcash/models/billing"
	"bandcash/models/dev"
	"bandcash/models/event"
	"bandcash/models/expense"
	"bandcash/models/group"
	"bandcash/models/health"
	"bandcash/models/home"
	"bandcash/models/member"
	"bandcash/models/sse"
)

func registerRoutes(e *echo.Echo) {
	const (
		termlyTermsURL   = "https://app.termly.io/policy-viewer/policy.html?policyUUID=3943ac6d-aba4-4da3-a9d3-c5f943607491"
		termlyPrivacyURL = "https://app.termly.io/policy-viewer/policy.html?policyUUID=b6e0c0b3-1823-464f-94a4-edc886a6a2a7"
		termlyCookiesURL = "https://app.termly.io/policy-viewer/policy.html?policyUUID=5e3aa2d8-185a-46be-8472-21459302efa4"
	)

	redirectTo := func(target string) echo.HandlerFunc {
		return func(c echo.Context) error {
			return c.Redirect(http.StatusFound, target)
		}
	}

	e.Static("/static", "static")
	e.File("/favicon.ico", "static/favicon.ico")
	e.GET("/health", health.Check)
	e.GET("/login", auth.LoginPageHandler)
	e.POST("/login", auth.LoginRequest, middleware.AuthBodyLimit, middleware.AuthRateLimit)
	e.GET("/login/verify", auth.VerifyMagicLink)
	e.DELETE("/session", auth.Logout)
	e.POST("/lemon_webhook", billingmodel.LemonWebhook)

	adminRoutes := e.Group("/admin", middleware.RequireAuth, middleware.RequireSuperadmin)
	adminRoutes.GET("", admin.Dashboard)
	adminRoutes.GET("/flags", admin.FlagsPage)
	adminRoutes.GET("/users", admin.UsersPage)
	adminRoutes.GET("/groups", admin.GroupsPage)
	adminRoutes.GET("/sessions", admin.SessionsPage)
	adminRoutes.POST("/flags/signup", admin.UpdateSignupFlag)
	adminRoutes.POST("/users/:userId/ban", admin.BanUser)
	adminRoutes.POST("/users/:userId/unban", admin.UnbanUser)
	adminRoutes.DELETE("/users/:id/sessions/:sessionid", admin.LogoutSession)
	adminRoutes.DELETE("/users/:id/sessions/", admin.LogoutAllUserSessions)

	grp := group.New()
	e.GET("/groups", grp.IndexPage, middleware.RequireAuth)
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth, middleware.RequireCanCreateGroup)
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth, middleware.RequireCanCreateGroup)

	groupUserRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	groupUserRoutes.GET("", grp.RootPage)
	groupUserRoutes.GET("/about", grp.AboutPage)
	groupUserRoutes.GET("/pending-payouts", grp.ToPayPage)
	groupUserRoutes.GET("/pending-incomes", grp.ToReceivePage)
	groupUserRoutes.GET("/recent-incomes", grp.RecentIncomePage)
	groupUserRoutes.GET("/recent-payouts", grp.RecentOutgoingPage)
	groupUserRoutes.GET("/users", grp.UsersPage)
	groupUserRoutes.GET("/users/:id", grp.UsersEntryPage)
	groupUserRoutes.POST("/leave", grp.LeaveGroup)

	groupAdminRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup, middleware.RequireAdmin)
	groupAdminRoutes.GET("/edit", grp.EditGroupPage)
	groupAdminRoutes.GET("/users/new", grp.UsersNewPage)
	groupAdminRoutes.GET("/users/:id/edit", grp.UserEditPage)
	groupAdminRoutes.PUT("", grp.UpdateGroup)
	groupAdminRoutes.DELETE("", grp.DeleteGroup)
	groupAdminRoutes.POST("/users", grp.AddViewer)
	groupAdminRoutes.DELETE("/users/:id", grp.DeleteUserEntry)
	groupAdminRoutes.PUT("/users/:id/admin", grp.PromoteViewerToAdmin)
	groupAdminRoutes.PUT("/users/:id/viewer", grp.DemoteAdminToViewer)
	groupAdminRoutes.PUT("/payments/events/:id/toggle-paid", grp.TogglePaymentEventPaid)
	groupAdminRoutes.POST("/payments/events/:id/paid_at", grp.UpdatePaymentEventPaidAt)
	groupAdminRoutes.PUT("/payments/participants/:eventId/:memberId/toggle-paid", grp.TogglePaymentParticipantPaid)
	groupAdminRoutes.POST("/payments/participants/:eventId/:memberId/paid_at", grp.UpdatePaymentParticipantPaidAt)
	groupAdminRoutes.PUT("/payments/expenses/:id/toggle-paid", grp.TogglePaymentExpensePaid)
	groupAdminRoutes.POST("/payments/expenses/:id/paid_at", grp.UpdatePaymentExpensePaidAt)

	groupOwnerRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup, middleware.RequireOwner)
	groupOwnerRoutes.PUT("/users/:id/transfer-owner", grp.TransferGroupOwnership)

	e.GET("/", home.Index)
	e.GET("/pricing", home.Pricing)
	e.GET("/terms", redirectTo(termlyTermsURL))
	e.GET("/privacy", redirectTo(termlyPrivacyURL))
	e.GET("/cookies", redirectTo(termlyCookiesURL))
	e.GET("/terms-and-conditions", redirectTo(termlyTermsURL))
	e.GET("/privacy-policy", redirectTo(termlyPrivacyURL))

	eventRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	eventRoutes.GET("/events", event.IndexPage)
	eventRoutes.GET("/overview", func(c echo.Context) error {
		groupID := c.Param("groupId")
		return c.Redirect(http.StatusMovedPermanently, "/groups/"+groupID+"/events")
	})
	eventRoutes.GET("/events/:id", event.ShowPage)
	eventRoutes.GET("/events/:id/members/:memberId/note", event.OpenParticipantNoteDialog)

	eventAdminRoutes := eventRoutes.Group("", middleware.RequireAdmin)
	eventAdminRoutes.GET("/events/new", event.NewEventPage)
	eventAdminRoutes.GET("/events/:id/edit", event.EditEventPage)
	eventAdminRoutes.GET("/events/:id/participant/edit", event.EditEventParticipantsPage)
	eventAdminRoutes.POST("/events", event.Create)
	eventAdminRoutes.POST("/events/:id", event.Update)
	eventAdminRoutes.POST("/events/:id/details", event.UpdateDetails)
	eventAdminRoutes.POST("/events/:id/paid", event.TogglePaid)
	eventAdminRoutes.GET("/events/:id/paid_at", event.OpenPaidAtPrompt)
	eventAdminRoutes.POST("/events/:id/paid_at", event.UpdatePaidAt)
	eventAdminRoutes.GET("/events/:id/members/:memberId/paid_at", event.OpenParticipantPaidAtDialog)
	eventAdminRoutes.POST("/events/:id/members/:memberId/paid_at", event.UpdateParticipantPaidAt)
	eventAdminRoutes.POST("/events/:id/members/:memberId/note", event.UpdateParticipantNote)
	eventAdminRoutes.POST("/events/:id/participants/draft", event.OpenParticipantsDraft)
	eventAdminRoutes.POST("/events/:id/participants/draft/rows", event.UpdateParticipantsDraftRows)
	eventAdminRoutes.PUT("/events/:id/participants", event.SaveParticipantsBulk)
	eventAdminRoutes.DELETE("/events/:id/participants/draft", event.CancelParticipantsDraft)
	eventAdminRoutes.DELETE("/events/:id", event.Destroy)
	eventAdminRoutes.POST("/events/:id/members/:memberId/paid", event.ToggleParticipantPaid)

	expenseRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	expenseRoutes.GET("/expenses", expense.IndexPage)
	expenseRoutes.GET("/expenses/:id", expense.ShowPage)

	expenseAdminRoutes := expenseRoutes.Group("", middleware.RequireAdmin)
	expenseAdminRoutes.GET("/expenses/new", expense.NewExpensePage)
	expenseAdminRoutes.GET("/expenses/:id/edit", expense.EditExpensePage)
	expenseAdminRoutes.POST("/expenses", expense.Create)
	expenseAdminRoutes.PUT("/expenses/:id", expense.Update)
	expenseAdminRoutes.DELETE("/expenses/:id", expense.Destroy)
	expenseAdminRoutes.PUT("/expenses/:id/toggle-paid", expense.TogglePaid)
	expenseAdminRoutes.GET("/expenses/:id/paid_at", expense.OpenPaidAtPrompt)
	expenseAdminRoutes.POST("/expenses/:id/paid_at", expense.UpdatePaidAt)

	memberRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	memberRoutes.GET("/members", member.Index)
	memberRoutes.GET("/members/:id", member.Show)

	memberAdminRoutes := memberRoutes.Group("", middleware.RequireAdmin)
	memberAdminRoutes.GET("/members/new", member.NewMemberPage)
	memberAdminRoutes.GET("/members/:id/edit", member.EditMemberPage)
	memberAdminRoutes.POST("/members", member.Create)
	memberAdminRoutes.PUT("/members/:id", member.Update)
	memberAdminRoutes.GET("/members/:id/events/:eventId/paid_at", member.OpenParticipantPaidAtDialog)
	memberAdminRoutes.POST("/members/:id/events/:eventId/paid_at", member.UpdateParticipantPaidAt)
	memberAdminRoutes.PUT("/members/:id/events/:eventId/toggle-paid", member.ToggleParticipantPaid)
	memberAdminRoutes.DELETE("/members/:id", member.Destroy)

	e.GET("/account", account.Index, middleware.RequireAuth)
	e.POST("/account/language", account.UpdateLanguage, middleware.RequireAuth)
	e.POST("/account/subscription/seats", account.UpdateSeats, middleware.RequireAuth)
	e.DELETE("/account/sessions/:id", account.LogoutSession, middleware.RequireAuth)
	e.DELETE("/account/sessions", account.LogoutAllOtherSessions, middleware.RequireAuth)

	e.GET("/sse", sse.SSEHandler())

	if utils.Env().AppEnv == "development" {
		devRoutes := e.Group("/dev")
		devRoutes.GET("", dev.DevPageHandler)
		devRoutes.POST("/spinner", dev.TestSpinner)
		devRoutes.POST("/multi-action/:action", dev.TestMultiAction)
		devRoutes.POST("/notifications/inline", dev.TestInline)
		devRoutes.POST("/notifications/test", dev.Test)
		devRoutes.GET("/emails/login", dev.PreviewLoginEmail)
		devRoutes.GET("/emails/invite", dev.PreviewInviteEmail)
		devRoutes.GET("/emails/invite-accepted", dev.PreviewInviteAcceptedEmail)
		devRoutes.GET("/emails/group-created", dev.PreviewGroupCreatedEmail)
		devRoutes.GET("/emails/role-upgraded", dev.PreviewRoleUpgradedEmail)
		devRoutes.GET("/emails/role-downgraded", dev.PreviewRoleDowngradedEmail)
		devRoutes.GET("/emails/access-removed", dev.PreviewAccessRemovedEmail)
		devRoutes.GET("/errors/link-invalid", dev.PreviewInvalidLinkErrorPage)
		devRoutes.GET("/errors/400", dev.PreviewBadRequestErrorPage)
		devRoutes.GET("/errors/403", dev.PreviewForbiddenErrorPage)
		devRoutes.GET("/errors/404", dev.PreviewNotFoundErrorPage)
		devRoutes.GET("/errors/500", dev.PreviewInternalErrorPage)
	}
}
