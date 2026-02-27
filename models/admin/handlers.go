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

	recentUsersRaw, err := db.Qry.ListRecentUsersWithBanStatus(c.Request().Context(), 10)
	if err != nil {
		slog.Error("admin.dashboard: failed to list users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	recentUsers := mapRecentUsers(recentUsersRaw)

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

func mapRecentUsers(rows []db.ListRecentUsersWithBanStatusRow) []RecentUserRow {
	users := make([]RecentUserRow, 0, len(rows))
	for _, row := range rows {
		users = append(users, RecentUserRow{
			ID:        row.ID,
			Email:     row.Email,
			CreatedAt: row.CreatedAt,
			IsBanned:  row.IsBanned != 0,
		})
	}
	return users
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

func (a *Admin) BanUser(c echo.Context) error {
	return a.setUserBanState(c, true)
}

func (a *Admin) UnbanUser(c echo.Context) error {
	return a.setUserBanState(c, false)
}

func (a *Admin) setUserBanState(c echo.Context, banned bool) error {
	userID := c.Param("userId")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == userID {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "admin.users.cannot_ban_self"))
		return a.patchRecentUsers(c)
	}

	if banned {
		err := db.Qry.BanUser(c.Request().Context(), db.BanUserParams{ID: utils.GenerateID("ban"), UserID: userID})
		if err != nil {
			slog.Error("admin.users.ban: failed to ban user", "user_id", userID, "err", err)
			utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "admin.users.ban_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "admin.users.banned"))
	} else {
		err := db.Qry.UnbanUser(c.Request().Context(), userID)
		if err != nil {
			slog.Error("admin.users.unban: failed to unban user", "user_id", userID, "err", err)
			utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "admin.users.unban_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "admin.users.unbanned"))
	}

	return a.patchRecentUsers(c)
}

func (a *Admin) patchRecentUsers(c echo.Context) error {
	notificationsHTML, err := utils.RenderComponentStringFor(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	recentUsersRaw, err := db.Qry.ListRecentUsersWithBanStatus(c.Request().Context(), 10)
	if err != nil {
		slog.Error("admin.users.patch: failed to list users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	recentUsers := mapRecentUsers(recentUsersRaw)

	usersHTML, err := utils.RenderComponentStringFor(c, RecentUsersTable(recentUsers))
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, usersHTML)
	}

	return c.NoContent(http.StatusOK)
}
