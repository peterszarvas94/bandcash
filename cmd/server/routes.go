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
	adminRoutes.POST("/flags/payments", admin.UpdatePaymentsFlag)
	adminRoutes.POST("/users/:userId/ban", admin.BanUser)
	adminRoutes.POST("/users/:userId/unban", admin.UnbanUser)
	adminRoutes.DELETE("/users/:id/sessions/:sessionid", admin.LogoutSession)
	adminRoutes.DELETE("/users/:id/sessions/", admin.LogoutAllUserSessions)

	grp := group.New()
	e.GET("/groups", grp.IndexPage, middleware.RequireAuth, middleware.RequireWithinSubscriptionLimit)
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth, middleware.RequireWithinSubscriptionLimit, middleware.RequireCanCreateGroup)
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth, middleware.RequireWithinSubscriptionLimit, middleware.RequireCanCreateGroup)

	groupUserRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	groupUserRoutes.GET("", grp.RootPage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/about", grp.AboutPage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/pending-payouts", grp.ToPayPage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/pending-incomes", grp.ToReceivePage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/recent-incomes", grp.RecentIncomePage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/recent-payouts", grp.RecentOutgoingPage, middleware.RequireWithinSubscriptionLimit)
	groupUserRoutes.GET("/users", grp.UsersPage)
	groupUserRoutes.GET("/users/:id", grp.UsersEntryPage)
	groupUserRoutes.POST("/leave", grp.LeaveGroup)

	groupAdminRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup, middleware.RequireAdmin)
	groupAdminRoutes.GET("/edit", grp.EditGroupPage, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.GET("/users/new", grp.UsersNewPage)
	groupAdminRoutes.GET("/users/:id/edit", grp.UserEditPage, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.PUT("", grp.UpdateGroup, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.DELETE("", grp.DeleteGroup, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.POST("/users", grp.AddViewer)
	groupAdminRoutes.DELETE("/users/:id", grp.DeleteUserEntry)
	groupAdminRoutes.PUT("/users/:id/admin", grp.PromoteViewerToAdmin)
	groupAdminRoutes.PUT("/users/:id/viewer", grp.DemoteAdminToViewer)
	groupAdminRoutes.PUT("/payments/events/:id/toggle-paid", grp.TogglePaymentEventPaid, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.POST("/payments/events/:id/paid_at", grp.UpdatePaymentEventPaidAt, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.PUT("/payments/participants/:eventId/:memberId/toggle-paid", grp.TogglePaymentParticipantPaid, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.POST("/payments/participants/:eventId/:memberId/paid_at", grp.UpdatePaymentParticipantPaidAt, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.PUT("/payments/expenses/:id/toggle-paid", grp.TogglePaymentExpensePaid, middleware.RequireWithinSubscriptionLimit)
	groupAdminRoutes.POST("/payments/expenses/:id/paid_at", grp.UpdatePaymentExpensePaidAt, middleware.RequireWithinSubscriptionLimit)

	e.GET("/", home.Index)
	e.GET("/pricing", home.Pricing)
	e.GET("/terms", redirectTo(termlyTermsURL))
	e.GET("/privacy", redirectTo(termlyPrivacyURL))
	e.GET("/cookies", redirectTo(termlyCookiesURL))
	e.GET("/terms-and-conditions", redirectTo(termlyTermsURL))
	e.GET("/privacy-policy", redirectTo(termlyPrivacyURL))

	eventRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	eventRoutes.GET("/events", event.IndexPage, middleware.RequireWithinSubscriptionLimit)
	eventRoutes.GET("/overview", func(c echo.Context) error {
		groupID := c.Param("groupId")
		return c.Redirect(http.StatusMovedPermanently, "/groups/"+groupID+"/events")
	})
	eventRoutes.GET("/events/:id", event.ShowPage, middleware.RequireWithinSubscriptionLimit)
	eventRoutes.GET("/events/:id/members/:memberId/note", event.OpenParticipantNoteDialog, middleware.RequireWithinSubscriptionLimit)

	eventAdminRoutes := eventRoutes.Group("", middleware.RequireAdmin)
	eventAdminRoutes.GET("/events/new", event.NewEventPage, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.GET("/events/:id/edit", event.EditEventPage, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.GET("/events/:id/participant/edit", event.EditEventParticipantsPage, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events", event.Create, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id", event.Update, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/details", event.UpdateDetails, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/paid", event.TogglePaid, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.GET("/events/:id/paid_at", event.OpenPaidAtPrompt, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/paid_at", event.UpdatePaidAt, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.GET("/events/:id/members/:memberId/paid_at", event.OpenParticipantPaidAtDialog, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/members/:memberId/paid_at", event.UpdateParticipantPaidAt, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/members/:memberId/note", event.UpdateParticipantNote, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/participants/draft", event.OpenParticipantsDraft, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/participants/draft/rows", event.UpdateParticipantsDraftRows, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.PUT("/events/:id/participants", event.SaveParticipantsBulk, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.DELETE("/events/:id/participants/draft", event.CancelParticipantsDraft, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.DELETE("/events/:id", event.Destroy, middleware.RequireWithinSubscriptionLimit)
	eventAdminRoutes.POST("/events/:id/members/:memberId/paid", event.ToggleParticipantPaid, middleware.RequireWithinSubscriptionLimit)

	expenseRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	expenseRoutes.GET("/expenses", expense.IndexPage, middleware.RequireWithinSubscriptionLimit)
	expenseRoutes.GET("/expenses/:id", expense.ShowPage, middleware.RequireWithinSubscriptionLimit)

	expenseAdminRoutes := expenseRoutes.Group("", middleware.RequireAdmin)
	expenseAdminRoutes.GET("/expenses/new", expense.NewExpensePage, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.GET("/expenses/:id/edit", expense.EditExpensePage, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.POST("/expenses", expense.Create, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.PUT("/expenses/:id", expense.Update, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.DELETE("/expenses/:id", expense.Destroy, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.PUT("/expenses/:id/toggle-paid", expense.TogglePaid, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.GET("/expenses/:id/paid_at", expense.OpenPaidAtPrompt, middleware.RequireWithinSubscriptionLimit)
	expenseAdminRoutes.POST("/expenses/:id/paid_at", expense.UpdatePaidAt, middleware.RequireWithinSubscriptionLimit)

	memberRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.RequireGroup)
	memberRoutes.GET("/members", member.Index, middleware.RequireWithinSubscriptionLimit)
	memberRoutes.GET("/members/:id", member.Show, middleware.RequireWithinSubscriptionLimit)

	memberAdminRoutes := memberRoutes.Group("", middleware.RequireAdmin)
	memberAdminRoutes.GET("/members/new", member.NewMemberPage, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.GET("/members/:id/edit", member.EditMemberPage, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.POST("/members", member.Create, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.PUT("/members/:id", member.Update, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.GET("/members/:id/events/:eventId/paid_at", member.OpenParticipantPaidAtDialog, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.POST("/members/:id/events/:eventId/paid_at", member.UpdateParticipantPaidAt, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.PUT("/members/:id/events/:eventId/toggle-paid", member.ToggleParticipantPaid, middleware.RequireWithinSubscriptionLimit)
	memberAdminRoutes.DELETE("/members/:id", member.Destroy, middleware.RequireWithinSubscriptionLimit)

	e.GET("/account", account.Index, middleware.RequireAuth)
	e.GET("/account/subscription", account.SubscriptionPageHandler, middleware.RequireAuth)
	e.GET("/account/language", account.LanguagePageHandler, middleware.RequireAuth)
	e.GET("/account/sessions", account.SessionsPageHandler, middleware.RequireAuth)
	e.GET("/over-limit", account.OverLimitPageHandler, middleware.RequireAuth)
	e.GET("/account/subscription/manage", account.ManageSubscription, middleware.RequireAuth)
	e.GET("/account/subscription/update-payment", account.UpdatePaymentMethod, middleware.RequireAuth)
	e.POST("/account/language", account.UpdateLanguage, middleware.RequireAuth)
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
