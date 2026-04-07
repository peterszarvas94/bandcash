package group

import (
	"github.com/labstack/echo/v4"

	"bandcash/internal/middleware"
)

func RegisterRoutes(e *echo.Echo) *Group {
	grp := New()

	// Group creation (requires auth)
	e.GET("/groups", grp.IndexPage, middleware.RequireAuth, middleware.WithDetailState)
	e.GET("/groups/new", grp.NewGroupPage, middleware.RequireAuth, middleware.WithDetailState)
	e.POST("/groups", grp.CreateGroup, middleware.RequireAuth, middleware.WithDetailState)

	// Group users pages for any group member
	usersRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup)
	usersRoutes.GET("", grp.RootPage)
	usersRoutes.GET("/about", grp.AboutPage)
	usersRoutes.GET("/pay", grp.PaymentsPage)
	usersRoutes.GET("/recent", grp.RecentPaymentsPage)
	usersRoutes.GET("/users", grp.UsersPage)
	usersRoutes.GET("/users/:id", grp.UsersEntryPage)
	usersRoutes.POST("/leave", grp.LeaveGroup)

	// User management (admin only)
	adminUsersRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup, middleware.RequireAdmin)
	adminUsersRoutes.GET("/edit", grp.EditGroupPage)
	adminUsersRoutes.GET("/users/new", grp.UsersNewPage)
	adminUsersRoutes.GET("/users/:id/edit", grp.UserEditPage)
	adminUsersRoutes.PUT("", grp.UpdateGroup)
	adminUsersRoutes.DELETE("", grp.DeleteGroup)
	adminUsersRoutes.POST("/users", grp.AddViewer)
	adminUsersRoutes.DELETE("/users/:id", grp.DeleteUserEntry)
	adminUsersRoutes.PUT("/users/:id/admin", grp.PromoteViewerToAdmin)
	adminUsersRoutes.PUT("/users/:id/viewer", grp.DemoteAdminToViewer)
	adminUsersRoutes.PUT("/pay/events/:id/toggle-paid", grp.TogglePaymentEventPaid)
	adminUsersRoutes.POST("/pay/events/:id/paid_at", grp.UpdatePaymentEventPaidAt)
	adminUsersRoutes.PUT("/pay/participants/:eventId/:memberId/toggle-paid", grp.TogglePaymentParticipantPaid)
	adminUsersRoutes.POST("/pay/participants/:eventId/:memberId/paid_at", grp.UpdatePaymentParticipantPaidAt)
	adminUsersRoutes.PUT("/pay/expenses/:id/toggle-paid", grp.TogglePaymentExpensePaid)
	adminUsersRoutes.POST("/pay/expenses/:id/paid_at", grp.UpdatePaymentExpensePaidAt)

	ownerRoutes := e.Group("/groups/:groupId", middleware.RequireAuth, middleware.WithDetailState, middleware.RequireGroup, middleware.RequireOwner)
	ownerRoutes.PUT("/users/:id/transfer-owner", grp.TransferGroupOwnership)

	return grp
}
