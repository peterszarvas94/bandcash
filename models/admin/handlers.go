package admin

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"bandcash/internal/flags"
	"bandcash/internal/utils"
	adminstore "bandcash/models/admin/data"
	authstore "bandcash/models/auth/data"
	shared "bandcash/models/shared"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"
)

type adminTabSignals struct {
	TabID string `json:"tab_id"`
}

func Dashboard(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/admin/flags")
}

func FlagsPage(c echo.Context) error {
	utils.EnsureTabID(c)

	userID := utils.GetUserID(c)
	_, err := authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	signupEnabled, err := flags.IsSignupEnabled(c.Request().Context())
	if err != nil {
		slog.Error("admin.flags: failed to read enable_signup flag", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "flags")},
		},
		Tab:             "flags",
		SignupEnabled:   signupEnabled,
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	return utils.RenderPage(c, AdminFlagsPage(data))
}

func UsersPage(c echo.Context) error {
	utils.EnsureTabID(c)

	userID := utils.GetUserID(c)
	_, err := authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := parseAdminUsersQuery(c)
	totalItems, err := adminstore.CountUsersTable(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.users: failed to count users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	userRows, err := adminstore.ListUsersTable(c.Request().Context(), query.Search, query.Sort, query.Dir, query.PageSize, int(query.Offset()))
	if err != nil {
		slog.Error("admin.users: failed to list users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	users := mapUsers(userRows)

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "users")},
		},
		Tab:             "users",
		Users:           users,
		UserQuery:       query,
		UserPager:       utils.BuildTablePagination(totalItems, query),
		UsersTable:      AdminUsersTableLayout(),
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	return utils.RenderPage(c, AdminUsersPage(data))
}

func GroupsPage(c echo.Context) error {
	utils.EnsureTabID(c)

	userID := utils.GetUserID(c)
	_, err := authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := parseAdminGroupsQuery(c)
	totalItems, err := adminstore.CountGroupsTable(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.groups: failed to count groups", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	groups, err := adminstore.ListGroupsTable(c.Request().Context(), query.Search, query.Sort, query.Dir, query.PageSize, int(query.Offset()))
	if err != nil {
		slog.Error("admin.groups: failed to list groups", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "groups")},
		},
		Tab:             "groups",
		Groups:          groups,
		GroupQuery:      query,
		GroupPager:      utils.BuildTablePagination(totalItems, query),
		GroupsTable:     AdminGroupsTableLayout(),
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	return utils.RenderPage(c, AdminGroupsPage(data))
}

func SessionsPage(c echo.Context) error {
	utils.EnsureTabID(c)

	userID := utils.GetUserID(c)
	_, err := authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := parseAdminSessionsQuery(c)
	totalItems, err := adminstore.CountSessionsTable(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.sessions: failed to count sessions", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	sessionRows, err := adminstore.ListSessionsTable(c.Request().Context(), query.Search, query.Sort, query.Dir, query.PageSize, int(query.Offset()))
	if err != nil {
		slog.Error("admin.sessions: failed to list sessions", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	sessions := mapSessions(sessionRows)

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "sessions")},
		},
		Tab:             "sessions",
		Sessions:        sessions,
		SessionQuery:    query,
		SessionPager:    utils.BuildTablePagination(totalItems, query),
		SessionsTable:   AdminSessionsTableLayout(),
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	return utils.RenderPage(c, AdminSessionsPage(data))
}

func UpdateSignupFlag(c echo.Context) error {
	signals := adminTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	var next bool
	switch c.QueryParam("value") {
	case "1", "true", "on":
		next = true
	case "0", "false", "off":
		next = false
	default:
		current, err := flags.IsSignupEnabled(c.Request().Context())
		if err != nil {
			slog.Error("admin.flags.update_signup: failed to read flag", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		next = !current
	}

	if err := flags.SetSignupEnabled(c.Request().Context(), next); err != nil {
		slog.Error("admin.flags.update_signup: failed to update flag", "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.flags.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.flags.updated"))
	notificationsHTML, err := utils.RenderHTMLForRequest(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	flagsHTML, err := utils.RenderHTMLForRequest(c, FlagsContent(next))
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, flagsHTML)
	}

	return c.NoContent(http.StatusOK)
}

func BanUser(c echo.Context) error {
	return setUserBanState(c, true)
}

func UnbanUser(c echo.Context) error {
	return setUserBanState(c, false)
}

func setUserBanState(c echo.Context, banned bool) error {
	signals := adminTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := c.Param("userId")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	currentUserID := utils.GetUserID(c)
	if currentUserID == userID {
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.users.cannot_ban_self"))
		return patchRecentUsers(c)
	}

	if banned {
		err := authstore.BanUser(c.Request().Context(), authstore.BanUserParams{ID: utils.GenerateID("ban"), UserID: userID})
		if err != nil {
			slog.Error("admin.users.ban: failed to ban user", "user_id", userID, "err", err)
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.users.ban_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.users.banned"))
	} else {
		err := authstore.UnbanUser(c.Request().Context(), userID)
		if err != nil {
			slog.Error("admin.users.unban: failed to unban user", "user_id", userID, "err", err)
			utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.users.unban_failed"))
			return c.NoContent(http.StatusInternalServerError)
		}
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.users.unbanned"))
	}

	return patchRecentUsers(c)
}

func LogoutSession(c echo.Context) error {
	signals := adminTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := c.Param("id")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	sessionID := c.Param("sessionid")
	if !utils.IsValidID(sessionID, "ses") {
		return c.NoContent(http.StatusBadRequest)
	}

	currentUserID := utils.GetUserID(c)
	currentSessionID := currentSessionIDFromCookie(c)

	if err := authstore.DeleteUserSession(c.Request().Context(), authstore.DeleteUserSessionParams{ID: sessionID, UserID: userID}); err != nil {
		slog.Error("admin.sessions.logout: failed to delete session", "session_id", sessionID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.sessions.logout_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.sessions.logged_out"))
	if userID == currentUserID && currentSessionID != "" && currentSessionID == sessionID {
		utils.ClearSessionCookie(c)
		if err := utils.SSEHub.Redirect(c, "/login"); err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		return c.NoContent(http.StatusOK)
	}

	return patchRecentSessions(c)
}

func LogoutAllUserSessions(c echo.Context) error {
	signals := adminTabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	userID := c.Param("id")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := authstore.DeleteAllUserSessions(c.Request().Context(), userID); err != nil {
		slog.Error("admin.sessions.logout_all: failed to delete sessions", "user_id", userID, "err", err)
		utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.sessions.logout_all_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, ctxi18n.T(c.Request().Context(), "admin.sessions.logged_out_all"))
	if userID == utils.GetUserID(c) {
		utils.ClearSessionCookie(c)
		if err := utils.SSEHub.Redirect(c, "/login"); err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		return c.NoContent(http.StatusOK)
	}

	return patchRecentSessions(c)
}

func currentSessionIDFromCookie(c echo.Context) string {
	cookie, err := c.Cookie(utils.SessionCookieName)
	if err != nil || cookie.Value == "" {
		return ""
	}
	session, err := authstore.GetUserSessionByToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return ""
	}
	return session.ID
}

func patchRecentUsers(c echo.Context) error {
	notificationsHTML, err := utils.RenderHTMLForRequest(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	userID := utils.GetUserID(c)
	_, err = authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := parseAdminUsersQuery(c)
	totalItems, err := adminstore.CountUsersTable(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.users.patch: failed to count users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	userRows, err := adminstore.ListUsersTable(c.Request().Context(), query.Search, query.Sort, query.Dir, query.PageSize, int(query.Offset()))
	if err != nil {
		slog.Error("admin.users.patch: failed to list users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	users := mapUsers(userRows)

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "users")},
		},
		Tab:             "users",
		Users:           users,
		UserQuery:       query,
		UserPager:       utils.BuildTablePagination(totalItems, query),
		UsersTable:      AdminUsersTableLayout(),
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	html, err := utils.RenderHTMLForRequest(c, AdminUsersPage(data))
	if err != nil {
		slog.Error("admin.users.patch: failed to render page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func patchRecentSessions(c echo.Context) error {
	notificationsHTML, err := utils.RenderHTMLForRequest(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	userID := utils.GetUserID(c)
	_, err = authstore.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := parseAdminSessionsQuery(c)
	totalItems, err := adminstore.CountSessionsTable(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.sessions.patch: failed to count sessions", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	sessionRows, err := adminstore.ListSessionsTable(c.Request().Context(), query.Search, query.Sort, query.Dir, query.PageSize, int(query.Offset()))
	if err != nil {
		slog.Error("admin.sessions.patch: failed to list sessions", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	sessions := mapSessions(sessionRows)

	data := DashboardData{
		Title: ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard"), Href: "/admin/flags"},
			{Label: adminTabLabel(c.Request().Context(), "sessions")},
		},
		Tab:             "sessions",
		Sessions:        sessions,
		SessionQuery:    query,
		SessionPager:    utils.BuildTablePagination(totalItems, query),
		SessionsTable:   AdminSessionsTableLayout(),
		IsAuthenticated: true,
		IsSuperAdmin:    true,
	}

	html, err := utils.RenderHTMLForRequest(c, AdminSessionsPage(data))
	if err != nil {
		slog.Error("admin.sessions.patch: failed to render page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func adminTabLabel(ctx context.Context, tab string) string {
	switch tab {
	case "users":
		return ctxi18n.T(ctx, "admin.tab.users")
	case "groups":
		return ctxi18n.T(ctx, "admin.tab.groups")
	case "sessions":
		return ctxi18n.T(ctx, "admin.tab.sessions")
	default:
		return ctxi18n.T(ctx, "admin.tab.flags")
	}
}

func mapSessions(rows []adminstore.AdminSessionTableRow) []AdminSessionRow {
	sessions := make([]AdminSessionRow, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, AdminSessionRow{ID: row.ID, UserID: row.UserID, UserEmail: row.UserEmail, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt})
	}
	return sessions
}

func mapUsers(rows []adminstore.AdminUserTableRow) []RecentUserRow {
	users := make([]RecentUserRow, 0, len(rows))
	for _, row := range rows {
		users = append(users, RecentUserRow{ID: row.ID, Email: row.Email, CreatedAt: row.CreatedAt, IsBanned: row.IsBanned != 0})
	}
	return users
}

func parseAdminUsersQuery(c echo.Context) utils.TableQuery {
	return utils.TableQuery{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "pageSize", utils.DefaultTablePageSize),
		Search:   strings.TrimSpace(c.QueryParam("q")),
		Sort:     c.QueryParam("sort"),
		Dir:      c.QueryParam("dir"),
		SortSet:  c.QueryParam("sort") != "",
	}
}

func parseAdminGroupsQuery(c echo.Context) utils.TableQuery {
	return utils.TableQuery{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "pageSize", utils.DefaultTablePageSize),
		Search:   strings.TrimSpace(c.QueryParam("q")),
		Sort:     c.QueryParam("sort"),
		Dir:      c.QueryParam("dir"),
		SortSet:  c.QueryParam("sort") != "",
	}
}

func parseAdminSessionsQuery(c echo.Context) utils.TableQuery {
	return utils.TableQuery{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "pageSize", utils.DefaultTablePageSize),
		Search:   strings.TrimSpace(c.QueryParam("q")),
		Sort:     c.QueryParam("sort"),
		Dir:      c.QueryParam("dir"),
		SortSet:  c.QueryParam("sort") != "",
	}
}

func parseIntParam(c echo.Context, name string, defaultVal int) int {
	val := c.QueryParam(name)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed < 1 {
		return defaultVal
	}
	return parsed
}
