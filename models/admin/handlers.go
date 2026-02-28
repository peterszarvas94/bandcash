package admin

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

	// Get active tab (default: overview)
	tab := c.QueryParam("tab")
	if tab != "flags" && tab != "users" && tab != "groups" {
		tab = "overview"
	}

	// Always get stats for overview
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

	signupEnabled, err := utils.IsSignupEnabled(c.Request().Context())
	if err != nil {
		slog.Error("admin.dashboard: failed to read enable_signup flag", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Prepare data for tabs
	data := DashboardData{
		Title:         ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs:   []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard")}},
		UserEmail:     user.Email,
		Tab:           tab,
		UsersCount:    usersCount,
		GroupsCount:   groupsCount,
		EventsCount:   eventsCount,
		MembersCount:  membersCount,
		SignupEnabled: signupEnabled,
	}

	// Load users if on users tab
	if tab == "users" {
		query := parseAdminUsersQuery(c)
		totalItems, err := db.Qry.CountUsersFiltered(c.Request().Context(), query.Search)
		if err != nil {
			slog.Error("admin.dashboard: failed to count users", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		query = utils.ClampPage(query, totalItems)

		// Fetch users based on sort and direction
		var usersRaw interface{}
		switch query.Sort {
		case "email":
			if query.Dir == "desc" {
				rows, err := db.Qry.ListUsersByEmailDescFiltered(c.Request().Context(), db.ListUsersByEmailDescFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list users", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Users = mapEmailDescUserRows(rows)
			} else {
				rows, err := db.Qry.ListUsersByEmailAscFiltered(c.Request().Context(), db.ListUsersByEmailAscFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list users", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Users = mapEmailAscUserRows(rows)
			}
		case "createdAt":
			if query.Dir == "asc" {
				rows, err := db.Qry.ListUsersByCreatedAscFiltered(c.Request().Context(), db.ListUsersByCreatedAscFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list users", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Users = mapCreatedAscUserRows(rows)
			} else {
				rows, err := db.Qry.ListUsersByCreatedDescFiltered(c.Request().Context(), db.ListUsersByCreatedDescFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list users", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Users = mapCreatedDescUserRows(rows)
			}
		default:
			// Default to createdAt desc
			rows, err := db.Qry.ListUsersByCreatedDescFiltered(c.Request().Context(), db.ListUsersByCreatedDescFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.dashboard: failed to list users", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			data.Users = mapCreatedDescUserRows(rows)
		}
		_ = usersRaw

		data.UserPager = utils.BuildTablePagination(totalItems, query)
		data.UserQuery = query
	}

	// Load groups if on groups tab
	if tab == "groups" {
		query := parseAdminGroupsQuery(c)
		totalItems, err := db.Qry.CountGroupsFiltered(c.Request().Context(), query.Search)
		if err != nil {
			slog.Error("admin.dashboard: failed to count groups", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		query = utils.ClampPage(query, totalItems)

		// Fetch groups based on sort and direction
		var groupsRaw interface{}
		switch query.Sort {
		case "name":
			if query.Dir == "desc" {
				rows, err := db.Qry.ListGroupsByNameDescFiltered(c.Request().Context(), db.ListGroupsByNameDescFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list groups", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Groups = rows
			} else {
				rows, err := db.Qry.ListGroupsByNameAscFiltered(c.Request().Context(), db.ListGroupsByNameAscFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list groups", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Groups = rows
			}
		case "createdAt":
			if query.Dir == "asc" {
				rows, err := db.Qry.ListGroupsByCreatedAscFiltered(c.Request().Context(), db.ListGroupsByCreatedAscFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list groups", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Groups = rows
			} else {
				rows, err := db.Qry.ListGroupsByCreatedDescFiltered(c.Request().Context(), db.ListGroupsByCreatedDescFilteredParams{
					Search: query.Search,
					Offset: query.Offset(),
					Limit:  int64(query.PageSize),
				})
				if err != nil {
					slog.Error("admin.dashboard: failed to list groups", "err", err)
					return c.NoContent(http.StatusInternalServerError)
				}
				data.Groups = rows
			}
		default:
			// Default to createdAt desc
			rows, err := db.Qry.ListGroupsByCreatedDescFiltered(c.Request().Context(), db.ListGroupsByCreatedDescFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.dashboard: failed to list groups", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			data.Groups = rows
		}
		_ = groupsRaw

		data.GroupPager = utils.BuildTablePagination(totalItems, query)
		data.GroupQuery = query
	}

	return utils.RenderComponent(c, DashboardPage(data))
}

func mapEmailDescUserRows(rows []db.ListUsersByEmailDescFilteredRow) []RecentUserRow {
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

func mapEmailAscUserRows(rows []db.ListUsersByEmailAscFilteredRow) []RecentUserRow {
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

func mapCreatedAscUserRows(rows []db.ListUsersByCreatedAscFilteredRow) []RecentUserRow {
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

func mapCreatedDescUserRows(rows []db.ListUsersByCreatedDescFilteredRow) []RecentUserRow {
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

func parseAdminUsersQuery(c echo.Context) utils.TableQuery {
	return utils.TableQuery{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "pageSize", 50),
		Search:   strings.TrimSpace(c.QueryParam("q")),
		Sort:     c.QueryParam("sort"),
		Dir:      c.QueryParam("dir"),
		SortSet:  c.QueryParam("sort") != "",
	}
}

func parseAdminGroupsQuery(c echo.Context) utils.TableQuery {
	return utils.TableQuery{
		Page:     parseIntParam(c, "page", 1),
		PageSize: parseIntParam(c, "pageSize", 50),
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
	// Patch notifications first
	notificationsHTML, err := utils.RenderComponentStringFor(c, shared.Notifications())
	if err == nil {
		_ = utils.SSEHub.PatchHTML(c, notificationsHTML)
	}

	// Re-fetch full dashboard data for users tab
	userID := middleware.GetUserID(c)
	user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	// Get stats for overview (always needed for full page render)
	usersCount, err := db.Qry.CountUsers(c.Request().Context())
	if err != nil {
		slog.Error("admin.users.patch: failed to count users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	groupsCount, err := db.Qry.CountGroups(c.Request().Context())
	if err != nil {
		slog.Error("admin.users.patch: failed to count groups", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	eventsCount, err := db.Qry.CountEvents(c.Request().Context())
	if err != nil {
		slog.Error("admin.users.patch: failed to count events", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	membersCount, err := db.Qry.CountMembers(c.Request().Context())
	if err != nil {
		slog.Error("admin.users.patch: failed to count members", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	signupEnabled, err := utils.IsSignupEnabled(c.Request().Context())
	if err != nil {
		slog.Error("admin.users.patch: failed to read enable_signup flag", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// Get current query parameters from the request
	query := parseAdminUsersQuery(c)

	// Re-fetch total count and users based on current query
	totalItems, err := db.Qry.CountUsersFiltered(c.Request().Context(), query.Search)
	if err != nil {
		slog.Error("admin.users.patch: failed to count filtered users", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	query = utils.ClampPage(query, totalItems)

	// Fetch users based on sort and direction
	var users []RecentUserRow
	switch query.Sort {
	case "email":
		if query.Dir == "desc" {
			rows, err := db.Qry.ListUsersByEmailDescFiltered(c.Request().Context(), db.ListUsersByEmailDescFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.users.patch: failed to list users", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			users = mapEmailDescUserRows(rows)
		} else {
			rows, err := db.Qry.ListUsersByEmailAscFiltered(c.Request().Context(), db.ListUsersByEmailAscFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.users.patch: failed to list users", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			users = mapEmailAscUserRows(rows)
		}
	case "createdAt":
		if query.Dir == "asc" {
			rows, err := db.Qry.ListUsersByCreatedAscFiltered(c.Request().Context(), db.ListUsersByCreatedAscFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.users.patch: failed to list users", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			users = mapCreatedAscUserRows(rows)
		} else {
			rows, err := db.Qry.ListUsersByCreatedDescFiltered(c.Request().Context(), db.ListUsersByCreatedDescFilteredParams{
				Search: query.Search,
				Offset: query.Offset(),
				Limit:  int64(query.PageSize),
			})
			if err != nil {
				slog.Error("admin.users.patch: failed to list users", "err", err)
				return c.NoContent(http.StatusInternalServerError)
			}
			users = mapCreatedDescUserRows(rows)
		}
	default:
		// Default to createdAt desc
		rows, err := db.Qry.ListUsersByCreatedDescFiltered(c.Request().Context(), db.ListUsersByCreatedDescFilteredParams{
			Search: query.Search,
			Offset: query.Offset(),
			Limit:  int64(query.PageSize),
		})
		if err != nil {
			slog.Error("admin.users.patch: failed to list users", "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		users = mapCreatedDescUserRows(rows)
	}

	// Build full DashboardData for complete page render
	data := DashboardData{
		Title:         ctxi18n.T(c.Request().Context(), "admin.title"),
		Breadcrumbs:   []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "admin.dashboard")}},
		UserEmail:     user.Email,
		Tab:           "users",
		UsersCount:    usersCount,
		GroupsCount:   groupsCount,
		EventsCount:   eventsCount,
		MembersCount:  membersCount,
		SignupEnabled: signupEnabled,
		Users:         users,
		UserQuery:     query,
		UserPager:     utils.BuildTablePagination(totalItems, query),
	}

	// Render and patch the full DashboardPage
	html, err := utils.RenderComponentStringFor(c, DashboardPage(data))
	if err != nil {
		slog.Error("admin.users.patch: failed to render page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)

	return c.NoContent(http.StatusOK)
}
