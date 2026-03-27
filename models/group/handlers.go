package group

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	ctxi18nlib "github.com/invopop/ctxi18n"
	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Group struct {
	model      *GroupModel
	usersModel *UsersModel
}

type UsersModel struct{}

var errAtLeastOneAdmin = errors.New("at least one admin required")

func (u *UsersModel) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec(utils.StandardTableQuerySpecParams{
		DefaultSort:  "createdAt",
		DefaultDir:   "desc",
		AllowedSorts: []string{"email", "role", "status", "createdAt"},
	})
}

func New() *Group {
	return &Group{
		model:      NewModel(),
		usersModel: &UsersModel{},
	}
}

type createGroupSignals struct {
	TabID    string `json:"tab_id"`
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type addViewerSignals struct {
	TabID      string           `json:"tab_id"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	FormData   struct {
		Email string `json:"email" validate:"required,email,max=320"`
		Role  string `json:"role" validate:"required,oneof=viewer admin"`
	} `json:"formData"`
}

type updateGroupSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	FormData   struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type deleteGroupSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type tabSignals struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

// NewGroupPage shows the form to create a new group
func (g *Group) NewGroupPage(c echo.Context) error {
	utils.EnsureTabID(c)
	userEmail := getUserEmail(c)
	data := NewGroupPageData{
		Title:       ctxi18n.T(c.Request().Context(), "groups.new_page_title"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"}, {Label: ctxi18n.T(c.Request().Context(), "groups.new")}},
		UserEmail:   userEmail,
	}
	return utils.RenderPage(c, GroupNewPage(data))
}

// EditGroupPage shows the form to edit a group name.
func (g *Group) EditGroupPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := EditGroupPageData{
		Title: ctxi18n.T(c.Request().Context(), "groups.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + group.ID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "groups.edit")},
		},
		UserEmail: userEmail,
		GroupID:   groupID,
		Group:     group,
	}

	return utils.RenderPage(c, GroupEditPage(data))
}

// CreateGroup handles group creation
func (g *Group) CreateGroup(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		err := utils.SSEHub.Redirect(c, "/auth/login")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	signals := createGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Name = strings.TrimSpace(signals.FormData.Name)
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.Notify(c, "error", errs["name"])
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	name := signals.FormData.Name

	// Create group
	group, err := db.Qry.CreateGroup(c.Request().Context(), db.CreateGroupParams{
		ID:          utils.GenerateID("grp"),
		Name:        name,
		AdminUserID: userID,
	})
	if err != nil {
		slog.Error("group: failed to create group", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to create group")
	}

	slog.Info("group: created", "group_id", group.ID, "name", group.Name, "admin", userID)
	if userEmail := getUserEmail(c); userEmail != "" {
		err = email.Email().SendGroupCreated(c.Request().Context(), userEmail, group.Name, group.ID, utils.Env().URL)
		if err != nil {
			slog.Warn("group.create: failed to send group created email", "group_id", group.ID, "err", err)
		}
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.created"))

	// Redirect to group events
	err = utils.SSEHub.Redirect(c, "/groups/"+group.ID+"/events")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// GroupsPage lists groups the user can access
func (g *Group) GroupsPage(c echo.Context) error {
	utils.EnsureTabID(c)
	userID := middleware.GetUserID(c)
	userEmail := getUserEmail(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	query := utils.ParseTableQuery(c, g.model)

	data, err := g.model.GetGroupsPageData(c.Request().Context(), userID, query)
	if err != nil {
		slog.Error("group: failed to load groups", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load groups")
	}

	data.Title = ctxi18n.T(c.Request().Context(), "groups.page_title")
	data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}}
	data.UserEmail = userEmail

	return utils.RenderPage(c, GroupsPage(data))
}

// GroupPage shows group details and admin actions.
func (g *Group) GroupPage(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	return c.Redirect(http.StatusFound, "/groups/"+groupID+"/events")
}

// UpdateGroup updates group name (admin only).
func (g *Group) UpdateGroup(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	signals := updateGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.SSEHub.PatchSignals(c, map[string]any{"errors": utils.WithErrors([]string{"name"}, errs)})
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	_, err := db.Qry.UpdateGroupName(c.Request().Context(), db.UpdateGroupNameParams{
		Name: signals.FormData.Name,
		ID:   groupID,
	})
	if err != nil {
		slog.Error("group.update: failed to update group", "group_id", groupID, "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.update_failed"))
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.updated"))

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/events")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// LeaveGroup removes viewer access for the current user.
func (g *Group) LeaveGroup(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := middleware.GetGroupID(c)
	userID := middleware.GetUserID(c)
	var err error
	if userID == "" {
		err = utils.SSEHub.Redirect(c, "/auth/login")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	_, err = db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.group_not_found"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if isAdminUser(c.Request().Context(), groupID, userID) {
		if err := g.removeAdminAccess(c.Request().Context(), groupID, userID); err != nil {
			if err == errAtLeastOneAdmin {
				utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.at_least_one_admin"))
			} else {
				slog.Error("group: failed to leave as admin", "group_id", groupID, "user_id", userID, "err", err)
				utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.leave_failed"))
			}
			err = utils.SSEHub.Redirect(c, "/groups")
			if err != nil {
				return c.NoContent(http.StatusInternalServerError)
			}
			return c.NoContent(http.StatusOK)
		}

		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.left"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	err = db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to leave", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.leave_failed"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.left"))
	err = utils.SSEHub.Redirect(c, "/groups")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// DeleteGroup removes the group and all data (admin only).
func (g *Group) DeleteGroup(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userID := middleware.GetUserID(c)
	signals := deleteGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	var err error
	if userID == "" {
		err = utils.SSEHub.Redirect(c, "/auth/login")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.group_not_found"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if group.AdminUserID != userID && !middleware.IsSuperadmin(c) {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.admin_required"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	if err := db.Qry.DeleteGroup(c.Request().Context(), groupID); err != nil {
		slog.Error("group: failed to delete", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.delete_failed"))
		err = utils.SSEHub.Redirect(c, "/groups")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.deleted"))
	if signals.Mode == "table" {
		query := utils.NormalizeTableQuery(signals.TableQuery, g.model.TableQuerySpec())
		data, err := g.model.GetGroupsPageData(c.Request().Context(), userID, query)
		if err != nil {
			slog.Error("group.delete: failed to load dashboard data", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		data.Title = ctxi18n.T(c.Request().Context(), "groups.page_title")
		data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}}
		data.UserEmail = getUserEmail(c)

		html, err := utils.RenderHTMLForRequest(c, GroupsPage(data))
		if err != nil {
			slog.Error("group.delete: failed to render dashboard", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	err = utils.SSEHub.Redirect(c, "/groups")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// UsersPage shows unified users table and invite form.
func (g *Group) UsersPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	data, err := g.usersPageData(c, groupID, c.QueryParams())
	if err != nil {
		slog.Error("group: failed to load users page", "group_id", groupID, "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load users")
	}

	return utils.RenderPage(c, GroupUsersPage(data))
}

// UsersEntryPage shows a user or invite row details based on ID prefix.
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

// UsersNewPage shows the form to invite a new viewer/admin.
func (g *Group) UsersNewPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
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
		UserEmail: userEmail,
		GroupID:   groupID,
		Group:     group,
	}

	return utils.RenderPage(c, GroupUsersNewPage(data))
}

// UserEditPage shows user role edit page.
func (g *Group) UserEditPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	userID := c.Param("id")
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	isAdminRole := isAdminUser(ctx, groupID, userID)
	if !isAdminRole {
		readerCount, readerErr := db.Qry.IsGroupReader(ctx, db.IsGroupReaderParams{UserID: userID, GroupID: groupID})
		if readerErr != nil || readerCount == 0 {
			return c.NoContent(http.StatusNotFound)
		}
	}

	user, err := db.Qry.GetUserByID(ctx, userID)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	row := GroupUserRow{Kind: "user", Status: "active", Email: user.Email, UserID: user.ID, Role: "viewer"}
	if isAdminRole {
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
		UserEmail: getUserEmail(c),
		GroupID:   groupID,
		Group:     group,
		UserRow:   row,
	}

	return utils.RenderPage(c, GroupUserEditPage(data))
}

// UserPage shows details for a user access row.
func (g *Group) UserPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	if !utils.IsValidID(userID, "usr") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_user_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	isAdminRole := isAdminUser(ctx, groupID, userID)
	if !isAdminRole {
		readerCount, readerErr := db.Qry.IsGroupReader(ctx, db.IsGroupReaderParams{UserID: userID, GroupID: groupID})
		if readerErr != nil || readerCount == 0 {
			return c.NoContent(http.StatusNotFound)
		}
	}

	user, err := db.Qry.GetUserByID(ctx, userID)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	row := GroupUserRow{
		Kind:   "user",
		Status: "active",
		Role: func() string {
			if isAdminRole {
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
		UserEmail:     getUserEmail(c),
		CurrentUserID: middleware.GetUserID(c),
		GroupID:       groupID,
		Group:         group,
		UserRow:       row,
		IsAdmin:       middleware.IsAdmin(c),
	}

	return utils.RenderPage(c, GroupUserPage(data))
}

// UserInvitePage shows details for a pending invite row.
func (g *Group) UserInvitePage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	inviteID := c.Param("inviteId")
	if inviteID == "" {
		inviteID = c.Param("id")
	}
	if !utils.IsValidID(inviteID, "mag") {
		return c.NoContent(http.StatusBadRequest)
	}

	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Error("group.users_invite_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	invites, err := db.Qry.ListGroupPendingInvites(ctx, sql.NullString{String: groupID, Valid: true})
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
		UserEmail: getUserEmail(c),
		GroupID:   groupID,
		Group:     group,
		UserRow:   row,
		IsAdmin:   middleware.IsAdmin(c),
	}

	return utils.RenderPage(c, GroupUserInvitePage(data))
}

func (g *Group) redirectUsersPage(c echo.Context, groupID, messageKey, errorKey string, status int) error {
	if messageKey != "" {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), errorKey))
	}
	err := utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(status)
}

func (g *Group) patchUsersPage(c echo.Context, groupID, messageKey, errorKey string) error {
	if messageKey != "" {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), errorKey))
	}
	data, err := g.usersPageData(c, groupID, queryValuesFromReferer(c))
	if err != nil {
		slog.Error("group: failed to load users patch data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	html, err := utils.RenderHTMLForRequest(c, GroupUsersPage(data))
	if err != nil {
		slog.Error("group: failed to render users page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

func (g *Group) patchUsersPageWithState(c echo.Context, groupID string, query utils.TableQuery, messageKey, errorKey string) error {
	if messageKey != "" {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), errorKey))
	}
	data, err := g.usersPageData(c, groupID, tableQueryValues(query))
	if err != nil {
		slog.Error("group: failed to load users patch data", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	html, err := utils.RenderHTMLForRequest(c, GroupUsersPage(data))
	if err != nil {
		slog.Error("group: failed to render users page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

// AddViewer invites a user to the group with selected user role.
func (g *Group) AddViewer(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	signals := addViewerSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Email = strings.ToLower(strings.TrimSpace(signals.FormData.Email))
	signals.FormData.Role = normalizeInviteRole(signals.FormData.Role)
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.Notify(c, "error", errs["email"])
		return c.NoContent(http.StatusUnprocessableEntity)
	}
	emailAddress := signals.FormData.Email
	inviteRole := signals.FormData.Role
	var err error

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.group_not_found")
	}

	// If user exists and already has access, short-circuit
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAddress)
	if err == nil {
		if group.AdminUserID == user.ID {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_admin", "")
		}

		adminCount, adminErr := db.Qry.IsGroupAdmin(c.Request().Context(), db.IsGroupAdminParams{
			UserID:  user.ID,
			GroupID: groupID,
		})
		if adminErr == nil && adminCount > 0 {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_admin", "")
		}

		count, err := db.Qry.IsGroupReader(c.Request().Context(), db.IsGroupReaderParams{
			UserID:  user.ID,
			GroupID: groupID,
		})
		if err == nil && count > 0 {
			if inviteRole == "admin" {
				err = db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{
					UserID:  user.ID,
					GroupID: groupID,
				})
				if err == nil {
					_, err = db.Qry.CreateGroupAdmin(c.Request().Context(), db.CreateGroupAdminParams{
						ID:      utils.GenerateID("gad"),
						UserID:  user.ID,
						GroupID: groupID,
					})
				}
				if err != nil {
					slog.Error("group: failed to promote viewer to admin", "group_id", groupID, "user_id", user.ID, "err", err)
					return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
				}
				if err := sendRoleChangeEmail(c.Request().Context(), user, group.Name, group.ID, "admin"); err != nil {
					slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", user.ID, "err", err)
				}
				return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.viewer_promoted", "")
			}
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_viewer", "")
		}
	}

	// Create invite magic link that does not expire.
	token := utils.GenerateID("tok")

	_, err = db.Qry.CreateInviteMagicLink(c.Request().Context(), db.CreateInviteMagicLinkParams{
		ID:         utils.GenerateID("mag"),
		Token:      token,
		Email:      emailAddress,
		GroupID:    sql.NullString{String: groupID, Valid: true},
		InviteRole: inviteRole,
	})
	if err != nil {
		slog.Error("group: failed to create invite link", "err", err)
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invite_failed")
	}

	err = email.Email().SendGroupInvitation(c.Request().Context(), emailAddress, group.Name, token, utils.Env().URL)
	if err != nil {
		slog.Error("group: failed to send invite email", "err", err)
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.send_failed")
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.invite_sent"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// RemoveViewer removes user access from the group.
func (g *Group) RemoveViewer(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	if !utils.IsValidID(userID, "usr") {
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	ctx := c.Request().Context()
	if isAdminUser(ctx, groupID, userID) {
		if err := g.removeAdminAccess(ctx, groupID, userID); err != nil {
			if err == errAtLeastOneAdmin {
				return g.redirectUsersPage(c, groupID, "", "groups.errors.at_least_one_admin", http.StatusConflict)
			}
			slog.Error("group: failed to remove admin access", "group_id", groupID, "user_id", userID, "err", err)
			return g.redirectUsersPage(c, groupID, "", "groups.errors.remove_failed", http.StatusInternalServerError)
		}
		notifyAccessRemoved(ctx, groupID, userID)
		return g.redirectUsersPage(c, groupID, "groups.messages.viewer_removed", "", http.StatusOK)
	}

	err := db.Qry.RemoveGroupReader(ctx, db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to remove viewer", "err", err)
		return g.redirectUsersPage(c, groupID, "", "groups.errors.remove_failed", http.StatusInternalServerError)
	}
	notifyAccessRemoved(ctx, groupID, userID)

	return g.redirectUsersPage(c, groupID, "groups.messages.viewer_removed", "", http.StatusOK)
}

func (g *Group) PromoteViewerToAdmin(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	ctx := c.Request().Context()
	if !utils.IsValidID(userID, "usr") {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invalid_user")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	if isAdminUser(ctx, groupID, userID) {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.already_admin", "")
		}
		return g.redirectUsersPage(c, groupID, "groups.messages.already_admin", "", http.StatusOK)
	}

	count, err := db.Qry.IsGroupReader(ctx, db.IsGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil || count == 0 {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	err = db.Qry.RemoveGroupReader(ctx, db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to remove viewer while promoting", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	_, err = db.Qry.CreateGroupAdmin(ctx, db.CreateGroupAdminParams{
		ID:      utils.GenerateID("gad"),
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to promote viewer", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.promote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.promote_failed", http.StatusInternalServerError)
	}

	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err == nil {
		if user, userErr := db.Qry.GetUserByID(ctx, userID); userErr == nil {
			if mailErr := sendRoleChangeEmail(ctx, user, group.Name, group.ID, "admin"); mailErr != nil {
				slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", userID, "err", mailErr)
			}
		}
	}

	if signals.Mode == "table" {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.viewer_promoted", "")
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.viewer_promoted"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users/"+userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

func (g *Group) DemoteAdminToViewer(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		userID = c.Param("id")
	}
	ctx := c.Request().Context()
	if !utils.IsValidID(userID, "usr") {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.invalid_user")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_user", http.StatusBadRequest)
	}
	if !isAdminUser(ctx, groupID, userID) {
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.demote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.demote_failed", http.StatusInternalServerError)
	}
	if err := g.demoteAdminToViewer(ctx, groupID, userID); err != nil {
		if err == errAtLeastOneAdmin {
			if signals.Mode == "table" {
				return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.at_least_one_admin")
			}
			return g.redirectUsersPage(c, groupID, "", "groups.errors.at_least_one_admin", http.StatusConflict)
		}
		slog.Error("group: failed to demote admin", "group_id", groupID, "user_id", userID, "err", err)
		if signals.Mode == "table" {
			return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "", "groups.errors.demote_failed")
		}
		return g.redirectUsersPage(c, groupID, "", "groups.errors.demote_failed", http.StatusInternalServerError)
	}
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err == nil {
		if user, userErr := db.Qry.GetUserByID(ctx, userID); userErr == nil {
			if mailErr := sendRoleChangeEmail(ctx, user, group.Name, group.ID, "viewer"); mailErr != nil {
				slog.Warn("group: failed to send role-change email", "group_id", groupID, "user_id", userID, "err", mailErr)
			}
		}
	}

	if signals.Mode == "table" {
		return g.patchUsersPageWithState(c, groupID, signals.TableQuery, "groups.messages.admin_demoted", "")
	}
	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.admin_demoted"))
	err = utils.SSEHub.Redirect(c, "/groups/"+groupID+"/users/"+userID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// CancelInvite removes a pending invitation from the group.
func (g *Group) CancelInvite(c echo.Context) error {
	signals := tabSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if !utils.SetTabID(c, signals.TabID) {
		return c.NoContent(http.StatusBadRequest)
	}

	groupID := middleware.GetGroupID(c)
	inviteID := c.Param("inviteId")
	if inviteID == "" {
		inviteID = c.Param("id")
	}
	if !utils.IsValidID(inviteID, "mag") {
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invalid_invite", http.StatusBadRequest)
	}

	err := db.Qry.DeleteGroupPendingInvite(c.Request().Context(), db.DeleteGroupPendingInviteParams{
		ID:      inviteID,
		GroupID: sql.NullString{String: groupID, Valid: true},
	})
	if err != nil {
		slog.Error("group: failed to cancel invite", "err", err)
		return g.redirectUsersPage(c, groupID, "", "groups.errors.invite_cancel_failed", http.StatusInternalServerError)
	}

	return g.redirectUsersPage(c, groupID, "groups.messages.invite_cancelled", "", http.StatusOK)
}

func (g *Group) DeleteUserEntry(c echo.Context) error {
	id := c.Param("id")
	if utils.IsValidID(id, "usr") {
		return g.RemoveViewer(c)
	}
	if utils.IsValidID(id, "mag") {
		return g.CancelInvite(c)
	}
	return c.NoContent(http.StatusBadRequest)
}

func (g *Group) usersPageData(c echo.Context, groupID string, values url.Values) (UsersPageData, error) {
	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return UsersPageData{}, err
	}
	query := parseTableQueryFromValues(values, g.usersModel)

	rows, err := g.buildUserRows(ctx, groupID)
	if err != nil {
		return UsersPageData{}, err
	}

	rows = filterUserRows(rows, query.Search)
	sortUserRows(rows, query)
	total := int64(len(rows))
	query = utils.ClampPage(query, total)
	rows = pageUserRows(rows, query)

	return UsersPageData{
		Title: ctxi18n.T(ctx, "groups.users_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + group.ID + "/events"},
			{Label: ctxi18n.T(ctx, "groups.users"), Href: "/groups/" + group.ID + "/users"},
		},
		UserEmail:     getUserEmail(c),
		CurrentUserID: middleware.GetUserID(c),
		Group:         group,
		UserRows:      rows,
		IsAdmin:       middleware.IsAdmin(c),
		Query:         query,
		Pager:         utils.BuildTablePagination(total, query),
		GroupID:       groupID,
		UsersTable:    utils.GroupUsersTableLayout(),
	}, nil
}

func queryValuesFromReferer(c echo.Context) url.Values {
	referer := c.Request().Referer()
	if referer == "" {
		return url.Values{}
	}
	u, err := url.Parse(referer)
	if err != nil {
		return url.Values{}
	}
	return u.Query()
}

func tableQueryValues(query utils.TableQuery) url.Values {
	values := url.Values{}
	values.Set("page", strconv.Itoa(query.Page))
	values.Set("pageSize", strconv.Itoa(query.PageSize))
	values.Set("q", query.Search)
	if query.SortSet && query.Sort != "" {
		values.Set("sort", query.Sort)
	}
	if query.Dir != "" {
		values.Set("dir", query.Dir)
	}
	return values
}

func (g *Group) buildUserRows(ctx context.Context, groupID string) ([]GroupUserRow, error) {
	admins, err := listGroupAdmins(ctx, groupID)
	if err != nil {
		return nil, err
	}
	adminIDs := make(map[string]struct{}, len(admins))
	rows := make([]GroupUserRow, 0, len(admins))
	for _, admin := range admins {
		adminIDs[admin.ID] = struct{}{}
		rows = append(rows, GroupUserRow{
			Kind:      "user",
			Status:    "active",
			Role:      "admin",
			Email:     admin.Email,
			UserID:    admin.ID,
			CreatedAt: admin.CreatedAt.Time,
		})
	}

	viewers, err := db.Qry.GetGroupReaders(ctx, groupID)
	if err != nil {
		return nil, err
	}
	for _, viewer := range viewers {
		if _, isAdmin := adminIDs[viewer.ID]; isAdmin {
			continue
		}
		rows = append(rows, GroupUserRow{
			Kind:      "user",
			Status:    "active",
			Role:      "viewer",
			Email:     viewer.Email,
			UserID:    viewer.ID,
			CreatedAt: viewer.CreatedAt.Time,
		})
	}

	invites, err := db.Qry.ListGroupPendingInvites(ctx, sql.NullString{String: groupID, Valid: true})
	if err != nil {
		return nil, err
	}
	for _, invite := range invites {
		createdAt := time.Time{}
		if invite.CreatedAt.Valid {
			createdAt = invite.CreatedAt.Time
		}
		rows = append(rows, GroupUserRow{
			Kind:      "invite",
			Status:    "pending",
			Role:      normalizeInviteRole(invite.InviteRole),
			Email:     invite.Email,
			InviteID:  invite.ID,
			CreatedAt: createdAt,
		})
	}

	return rows, nil
}

func filterUserRows(rows []GroupUserRow, search string) []GroupUserRow {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return rows
	}
	filtered := make([]GroupUserRow, 0, len(rows))
	for _, row := range rows {
		if strings.Contains(strings.ToLower(row.Email), search) || strings.Contains(strings.ToLower(row.Role), search) || strings.Contains(strings.ToLower(row.Status), search) {
			filtered = append(filtered, row)
		}
	}
	return filtered
}

func sortUserRows(rows []GroupUserRow, query utils.TableQuery) {
	less := func(i, j int) bool {
		a := rows[i]
		b := rows[j]
		switch query.Sort {
		case "email":
			return strings.ToLower(a.Email) < strings.ToLower(b.Email)
		case "role":
			return a.Role < b.Role
		case "status":
			return a.Status < b.Status
		default:
			if a.CreatedAt.Equal(b.CreatedAt) {
				return strings.ToLower(a.Email) < strings.ToLower(b.Email)
			}
			return a.CreatedAt.Before(b.CreatedAt)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if query.Dir == "desc" {
			return !less(i, j)
		}
		return less(i, j)
	})
}

func pageUserRows(rows []GroupUserRow, query utils.TableQuery) []GroupUserRow {
	if len(rows) == 0 {
		return rows
	}
	start := int(query.Offset())
	if start >= len(rows) {
		return []GroupUserRow{}
	}
	end := start + query.PageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[start:end]
}

func listGroupAdmins(ctx context.Context, groupID string) ([]db.User, error) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	primary, err := db.Qry.GetUserByID(ctx, group.AdminUserID)
	if err != nil {
		return nil, err
	}
	admins := []db.User{primary}
	seen := map[string]struct{}{primary.ID: {}}
	adminIDs, err := db.Qry.ListGroupAdminUserIDs(ctx, groupID)
	if err != nil {
		return nil, err
	}
	for _, adminID := range adminIDs {
		if _, ok := seen[adminID]; ok {
			continue
		}
		admin, userErr := db.Qry.GetUserByID(ctx, adminID)
		if userErr != nil {
			continue
		}
		admins = append(admins, admin)
		seen[admin.ID] = struct{}{}
	}
	return admins, nil
}

func isAdminUser(ctx context.Context, groupID, userID string) bool {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return false
	}
	if group.AdminUserID == userID {
		return true
	}
	count, err := db.Qry.IsGroupAdmin(ctx, db.IsGroupAdminParams{UserID: userID, GroupID: groupID})
	return err == nil && count > 0
}

func (g *Group) demoteAdminToViewer(ctx context.Context, groupID, userID string) error {
	if err := g.removeAdminAccess(ctx, groupID, userID); err != nil {
		return err
	}
	count, err := db.Qry.IsGroupReader(ctx, db.IsGroupReaderParams{UserID: userID, GroupID: groupID})
	if err == nil && count > 0 {
		return nil
	}
	_, err = db.Qry.CreateGroupReader(ctx, db.CreateGroupReaderParams{ID: utils.GenerateID("grd"), UserID: userID, GroupID: groupID})
	return err
}

func (g *Group) removeAdminAccess(ctx context.Context, groupID, userID string) error {
	adminUsers, err := listGroupAdmins(ctx, groupID)
	if err != nil {
		return err
	}
	if len(adminUsers) <= 1 {
		return errAtLeastOneAdmin
	}
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group.AdminUserID == userID {
		replacement := ""
		for _, admin := range adminUsers {
			if admin.ID != userID {
				replacement = admin.ID
				break
			}
		}
		if replacement == "" {
			return errAtLeastOneAdmin
		}
		if err := db.Qry.UpdateGroupAdmin(ctx, db.UpdateGroupAdminParams{AdminUserID: replacement, ID: groupID}); err != nil {
			return err
		}
	}
	_ = db.Qry.RemoveGroupAdmin(ctx, db.RemoveGroupAdminParams{UserID: userID, GroupID: groupID})
	_ = db.Qry.RemoveGroupReader(ctx, db.RemoveGroupReaderParams{UserID: userID, GroupID: groupID})
	return nil
}

func sendRoleChangeEmail(ctx context.Context, user db.User, groupName, groupID, role string) error {
	baseURL := utils.Env().URL
	mailCtx := ctx
	if user.PreferredLang != "" {
		if localizedCtx, err := ctxi18nlib.WithLocale(ctx, user.PreferredLang); err == nil {
			mailCtx = localizedCtx
		}
	}
	if role == "admin" {
		return email.Email().SendRoleUpgradedToAdmin(mailCtx, user.Email, groupName, groupID, baseURL)
	}
	return email.Email().SendRoleDowngradedToViewer(mailCtx, user.Email, groupName, groupID, baseURL)
}

func notifyAccessRemoved(ctx context.Context, groupID, userID string) {
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		slog.Warn("group: failed to load group for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}
	user, err := db.Qry.GetUserByID(ctx, userID)
	if err != nil {
		slog.Warn("group: failed to load user for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}
	admins, err := listGroupAdmins(ctx, groupID)
	if err != nil {
		slog.Warn("group: failed to list admins for access-removed email", "group_id", groupID, "user_id", userID, "err", err)
		return
	}

	adminEmails := make([]string, 0, len(admins))
	for _, admin := range admins {
		adminEmails = append(adminEmails, admin.Email)
	}

	mailCtx := ctx
	if user.PreferredLang != "" {
		if localizedCtx, localeErr := ctxi18nlib.WithLocale(ctx, user.PreferredLang); localeErr == nil {
			mailCtx = localizedCtx
		}
	}

	if err := email.Email().SendAccessRemoved(mailCtx, user.Email, group.Name, adminEmails, utils.Env().URL); err != nil {
		slog.Warn("group: failed to send access-removed email", "group_id", groupID, "user_id", userID, "err", err)
	}
}

func parseTableQueryFromValues(values url.Values, queryable utils.Queryable) utils.TableQuery {
	query := utils.TableQuery{
		Page:     parsePositiveInt(values.Get("page"), 1),
		PageSize: parsePositiveInt(values.Get("pageSize"), utils.DefaultTablePageSize),
		Search:   strings.TrimSpace(values.Get("q")),
		Sort:     values.Get("sort"),
		Dir:      values.Get("dir"),
		SortSet:  values.Get("sort") != "",
	}

	return utils.NormalizeTableQuery(query, queryable.TableQuerySpec())
}

func parsePositiveInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func normalizeInviteRole(role string) string {
	if strings.EqualFold(strings.TrimSpace(role), "admin") {
		return "admin"
	}
	return "viewer"
}

func getUserEmail(c echo.Context) string {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return ""
	}
	user, err := db.Qry.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return ""
	}
	return user.Email
}

func (g *Group) groupPageData(c echo.Context, groupID string) (GroupPageData, error) {
	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupPageData{}, err
	}

	admin, err := db.Qry.GetUserByID(ctx, group.AdminUserID)
	if err != nil {
		return GroupPageData{}, err
	}

	totals, err := utils.CalculateGroupTotals(ctx, groupID)
	if err != nil {
		return GroupPageData{}, err
	}

	return GroupPageData{
		Title:          "Bandcash - " + group.Name,
		Breadcrumbs:    []utils.Crumb{{Label: ctxi18n.T(ctx, "groups.title"), Href: "/groups"}, {Label: group.Name, Href: "/groups/" + groupID + "/events"}, {Label: ctxi18n.T(ctx, "nav.overview")}},
		UserEmail:      getUserEmail(c),
		Group:          group,
		Admin:          admin,
		Income:         totals.TotalEventAmount,
		IncomePaid:     totals.EventPaid,
		IncomeUnpaid:   totals.EventUnpaid,
		Payouts:        totals.TotalPayoutAmount,
		PayoutsPaid:    totals.PayoutPaid,
		PayoutsUnpaid:  totals.PayoutUnpaid,
		Expenses:       totals.TotalExpenseAmount,
		ExpensesPaid:   totals.ExpensePaid,
		ExpensesUnpaid: totals.ExpenseUnpaid,
		Leftover:       totals.TotalLeftover,
		IsAdmin:        middleware.IsAdmin(c),
	}, nil
}
