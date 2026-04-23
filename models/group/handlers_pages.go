package group

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	internalbilling "bandcash/internal/billing"
	"bandcash/internal/utils"
	authstore "bandcash/models/auth/data"
	groupstore "bandcash/models/group/data"
)

func (g *Group) NewGroupPage(c echo.Context) error {
	utils.EnsureTabID(c)
	data := NewGroupPageData{
		Title:           ctxi18n.T(c.Request().Context(), "groups.new_page_title"),
		Breadcrumbs:     []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"}, {Label: ctxi18n.T(c.Request().Context(), "groups.new")}},
		Signals: map[string]any{
			"formData":    map[string]any{"name": ""},
			"errors":      map[string]any{"name": ""},
			"groupCreate": map[string]any{"limitReached": false},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}
	return utils.RenderPage(c, GroupNewPage(data))
}

func (g *Group) EditGroupPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := EditGroupPageData{
		Title: ctxi18n.T(c.Request().Context(), "groups.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + group.ID + "/about"},
			{Label: ctxi18n.T(c.Request().Context(), "groups.edit")},
		},
		GroupID: groupID,
		Group:   group,
		Signals: map[string]any{
			"formData": map[string]any{"name": group.Name},
			"errors":   map[string]any{"name": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}

	return utils.RenderPage(c, GroupEditPage(data))
}

func (g *Group) IndexPage(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := utils.GetUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/login")
	}

	query := utils.ParseTableQuery(c, g.model)

	data, err := g.model.GetGroupsPageData(c.Request().Context(), userID, query)
	if err != nil {
		slog.Error("group: failed to load groups", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load groups")
	}

	data.Title = ctxi18n.T(c.Request().Context(), "groups.page_title")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}}
	data.Signals = nil
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	if state, err := internalbilling.CurrentAccessState(c.Request().Context(), userID); err == nil {
		data.RemainingSlots = internalbilling.RemainingGroupSlots(state)
	}

	return utils.RenderPage(c, GroupIndexPage(data))
}

func (g *Group) RootPage(c echo.Context) error {
	groupID := utils.GetGroupID(c)
	return c.Redirect(http.StatusFound, "/groups/"+groupID+"/events")
}

func (g *Group) AboutPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	data, err := g.groupPageData(c, groupID)
	if err != nil {
		slog.Error("group.about: failed to load data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.RenderPage(c, GroupAboutPage(data))
}

func (g *Group) ToPayPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: toPayPaymentsTableQuerySpec()})

	data, err := g.toPayPageData(c, groupID, query)
	if err != nil {
		slog.Error("group.to_pay: failed to load data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.RenderPage(c, GroupToPayPage(data))
}

func (g *Group) ToReceivePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: toReceivePaymentsTableQuerySpec()})

	data, err := g.toReceivePageData(c, groupID, query)
	if err != nil {
		slog.Error("group.to_receive: failed to load data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.RenderPage(c, GroupToReceivePage(data))
}

func (g *Group) RecentIncomePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: recentIncomePaymentsTableQuerySpec()})

	data, err := g.recentIncomePageData(c, groupID, query)
	if err != nil {
		slog.Error("group.recent_income: failed to load data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.RenderPage(c, GroupRecentIncomePage(data))
}

func (g *Group) RecentOutgoingPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: recentOutgoingPaymentsTableQuerySpec()})

	data, err := g.recentOutgoingPageData(c, groupID, query)
	if err != nil {
		slog.Error("group.recent_outgoing: failed to load data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return utils.RenderPage(c, GroupRecentOutgoingPage(data))
}

func (g *Group) UsersPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	data, err := g.usersPageData(c, groupID, c.QueryParams())
	if err != nil {
		slog.Error("group: failed to load users page", "group_id", groupID, "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load users")
	}

	return utils.RenderPage(c, GroupUsersPage(data))
}

func (g *Group) UsersEntryPage(c echo.Context) error {
	id := c.Param("id")
	if utils.IsValidID(id, "usr") {
		return g.UserPage(c)
	}
	if utils.IsValidID(id, "mag") {
		return g.UserInvitePage(c)
	}
	return c.NoContent(http.StatusBadRequest)
}

func (g *Group) UsersNewPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group.users_new_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := UsersNewPageData{
		Title: ctxi18n.T(c.Request().Context(), "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "groups.users"), Href: "/groups/" + groupID + "/users"},
			{Label: ctxi18n.T(c.Request().Context(), "groups.invite_user")},
		},
		GroupID: groupID,
		Group:   group,
		Signals: map[string]any{
			"formData": map[string]any{"email": "", "role": "viewer"},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}

	return utils.RenderPage(c, GroupUsersNewPage(data))
}

func (g *Group) UserEditPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	userID := c.Param("id")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	role, roleErr := getGroupAccessRole(ctx, groupID, userID)
	if roleErr != nil {
		return c.NoContent(http.StatusNotFound)
	}

	user, err := authstore.GetUserByID(ctx, userID)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	row := GroupUserRow{Kind: "user", Status: "active", Email: user.Email, UserID: user.ID, Role: "viewer"}
	switch role {
	case "owner":
		row.Role = "owner"
	case "admin":
		row.Role = "admin"
	}

	data := UserEditPageData{
		Title: ctxi18n.T(ctx, "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.users"), Href: "/groups/" + groupID + "/users"},
			{Label: user.Email, Href: "/groups/" + groupID + "/users/" + user.ID},
			{Label: ctxi18n.T(ctx, "actions.edit")},
		},
		GroupID:         groupID,
		Group:           group,
		UserRow:         row,
		Signals:         map[string]any{"formData": map[string]any{"role": row.Role}},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}

	return utils.RenderPage(c, GroupUserEditPage(data))
}

func (g *Group) UserPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_user_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	role, roleErr := getGroupAccessRole(ctx, groupID, userID)
	if roleErr != nil {
		return c.NoContent(http.StatusNotFound)
	}

	user, err := authstore.GetUserByID(ctx, userID)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	row := GroupUserRow{
		Kind:   "user",
		Status: "active",
		Role: func() string {
			if role == "owner" {
				return "owner"
			}
			if role == "admin" {
				return "admin"
			}
			return "viewer"
		}(),
		Email:  user.Email,
		UserID: user.ID,
		CreatedAt: func() time.Time {
			if user.CreatedAt.Valid {
				return user.CreatedAt.Time
			}
			return time.Time{}
		}(),
	}

	data := UserPageData{
		Title: ctxi18n.T(ctx, "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.users"), Href: "/groups/" + groupID + "/users"},
			{Label: user.Email},
		},
		CurrentUserID:   utils.GetUserID(c),
		GroupID:         groupID,
		Group:           group,
		UserRow:         row,
		IsAdmin:         utils.IsAdmin(c),
		Signals:         nil,
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}

	return utils.RenderPage(c, GroupUserPage(data))
}

func (g *Group) UserInvitePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	inviteID := c.Param("inviteId")
	if inviteID == "" {
		inviteID = c.Param("id")
	}
	if !utils.IsValidID(inviteID, "mag") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := groupstore.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_invite_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	invites, err := groupstore.ListGroupPendingInvites(ctx, sql.NullString{String: groupID, Valid: true})
	if err != nil {
		slog.Error("group.users_invite_page: failed to load invites", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var row GroupUserRow
	found := false
	for _, invite := range invites {
		if invite.ID != inviteID {
			continue
		}
		createdAt := time.Time{}
		if invite.CreatedAt.Valid {
			createdAt = invite.CreatedAt.Time
		}
		row = GroupUserRow{
			Kind:      "invite",
			Status:    "pending",
			Role:      normalizeInviteRole(invite.InviteRole),
			Email:     invite.Email,
			InviteID:  invite.ID,
			CreatedAt: createdAt,
		}
		found = true
		break
	}
	if !found {
		return c.NoContent(http.StatusNotFound)
	}

	data := UserInvitePageData{
		Title: ctxi18n.T(ctx, "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.users"), Href: "/groups/" + groupID + "/users"},
			{Label: row.Email},
		},
		GroupID:         groupID,
		Group:           group,
		UserRow:         row,
		IsAdmin:         utils.IsAdmin(c),
		Signals:         nil,
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}

	return utils.RenderPage(c, GroupUserInvitePage(data))
}
