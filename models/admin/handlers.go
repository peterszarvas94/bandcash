package admin

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
	shared "bandcash/models/shared"
)

type Admin struct{}

func New() *Admin {
	return &Admin{}
}

func (a *Admin) Dashboard(c echo.Context) error {
	utils.EnsureClientID(c)

	userID := middleware.GetUserID(c)
	user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	usersCount, err := db.Qry.CountUsers(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to count users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	groupsCount, err := db.Qry.CountGroups(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to count groups", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	eventsCount, err := db.Qry.CountEvents(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to count events", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	membersCount, err := db.Qry.CountMembers(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to count members", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	recentUsers, err := db.Qry.ListRecentUsers(c.Request().Context(), 10)
	if err != nil {
		slog.Error("admin.dashboard: failed to list users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	recentGroups, err := db.Qry.ListRecentGroups(c.Request().Context(), 10)
	if err != nil {
		slog.Error("admin.dashboard: failed to list groups", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	signupEnabled, err := utils.IsSignupEnabled(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to read enable_signup flag", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := DashboardData{
		Title:         ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs:   []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard")}},
		UserEmail:     user.Email,
		UsersCount:    usersCount,
		GroupsCount:   groupsCount,
		EventsCount:   eventsCount,
		MembersCount:  membersCount,
		SignupEnabled: signupEnabled,
		RecentUsers:   recentUsers,
		RecentGroups:  recentGroups,
	}

	return utils.RenderComponent(c, DashboardPage(data))
}

func (a *Admin) UpdateSignupFlag(c echo.Context) error {
	var next bool
	switch c.QueryParam("value") {
	case "1", "true", "on":
		next = true
	case "0", "false", "off":
		next = false
	default:
		current, err := utils.IsSignupEnabled(c.Request().Context())
		if err != nil {
			slog.Error("admin.flags.update_signup: failed to read flag", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		next = !current
	}

	err := utils.SetSignupEnabled(c.Request().Context(), next)
	if err != nil {
		slog.Error("admin.flags.update_signup: failed to update flag", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "admin.flags.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "admin.flags.updated"))
	notificationsHTML, err := utils.RenderComponentStringFor(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	flagsHTML, err := utils.RenderComponentStringFor(c, FlagsContent(next))
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, flagsHTML)
	}

	return c.NoContent(http.StatusOK)
}
