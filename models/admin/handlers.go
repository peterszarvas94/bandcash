package admin

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
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

	data := DashboardData{
		Title:        ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs:  []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard")}},
		UserEmail:    user.Email,
		UsersCount:   usersCount,
		GroupsCount:  groupsCount,
		EventsCount:  eventsCount,
		MembersCount: membersCount,
		RecentUsers:  recentUsers,
		RecentGroups: recentGroups,
	}

	return utils.RenderComponent(c, DashboardPage(data))
}
