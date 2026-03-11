package group

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"
	"github.com/starfederation/datastar-go/datastar"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Group struct {
	model               *GroupModel
	viewersModel        *ViewersModel
	adminsModel         *AdminsModel
	pendingInvitesModel *PendingInvitesModel
}

type ViewersModel struct{}
type AdminsModel struct{}
type PendingInvitesModel struct{}

const (
	viewersTabViewers = "viewers"
	viewersTabAdmins  = "admins"
	viewersTabPending = "pending"
)

func (v *ViewersModel) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec("email", "asc", "email")
}

func (a *AdminsModel) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec("email", "asc", "email")
}

func (p *PendingInvitesModel) TableQuerySpec() utils.TableQuerySpec {
	return utils.StandardTableQuerySpec("createdAt", "desc", "email", "createdAt")
}

func New() *Group {
	return &Group{
		model:               NewModel(),
		viewersModel:        &ViewersModel{},
		adminsModel:         &AdminsModel{},
		pendingInvitesModel: &PendingInvitesModel{},
	}
}

type createGroupSignals struct {
	FormData struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type addViewerSignals struct {
	FormData struct {
		Email string `json:"email" validate:"required,email,max=320"`
	} `json:"formData"`
}

type updateGroupSignals struct {
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	FormData   struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	} `json:"formData"`
}

type deleteGroupSignals struct {
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

// NewGroupPage shows the form to create a new group
func (g *Group) NewGroupPage(c echo.Context) error {
	utils.EnsureClientID(c)
	userEmail := getUserEmail(c)
	data := NewGroupPageData{
		Title:       ctxi18n.T(c.Request().Context(), "groups.new_page_title"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/dashboard"}, {Label: ctxi18n.T(c.Request().Context(), "groups.new")}},
		UserEmail:   userEmail,
	}
	return utils.RenderPage(c, GroupNewPage(data))
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

	// Redirect to group overview
	err = utils.SSEHub.Redirect(c, "/groups/"+group.ID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// GroupsPage lists groups the user can access
func (g *Group) GroupsPage(c echo.Context) error {
	utils.EnsureClientID(c)
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
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	data, err := g.groupPageData(c, groupID)
	if err != nil {
		slog.Error("group.page: failed to load data", "group_id", groupID, "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load group")
	}

	return utils.RenderPage(c, GroupPage(data))
}

// UpdateGroup updates group name (admin only).
func (g *Group) UpdateGroup(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userID := middleware.GetUserID(c)

	signals := updateGroupSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
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
	if signals.Mode == "table" {
		query := utils.NormalizeTableQuery(signals.TableQuery, g.model.TableQuerySpec())
		data, err := g.model.GetGroupsPageData(c.Request().Context(), userID, query)
		if err != nil {
			slog.Error("group.update: failed to load dashboard data", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		data.Title = ctxi18n.T(c.Request().Context(), "groups.page_title")
		data.Breadcrumbs = []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}}
		data.UserEmail = getUserEmail(c)

		html, err := utils.RenderHTMLForRequest(c, GroupsPage(data))
		if err != nil {
			slog.Error("group.update: failed to render dashboard", "group_id", groupID, "err", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		utils.SSEHub.PatchHTML(c, html)
		return c.NoContent(http.StatusOK)
	}

	err = utils.SSEHub.Redirect(c, "/groups/"+groupID)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}

// LeaveGroup removes viewer access for the current user.
func (g *Group) LeaveGroup(c echo.Context) error {
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

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.group_not_found"))
		err = utils.SSEHub.Redirect(c, "/dashboard")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if group.AdminUserID == userID {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.admin_cannot_leave"))
		err = utils.SSEHub.Redirect(c, "/dashboard")
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
		err = utils.SSEHub.Redirect(c, "/dashboard")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), "groups.messages.left"))
	err = utils.SSEHub.Redirect(c, "/dashboard")
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
		signals = deleteGroupSignals{}
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
		err = utils.SSEHub.Redirect(c, "/dashboard")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if group.AdminUserID != userID && !middleware.IsSuperadmin(c) {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.admin_required"))
		err = utils.SSEHub.Redirect(c, "/dashboard")
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	if err := db.Qry.DeleteGroup(c.Request().Context(), groupID); err != nil {
		slog.Error("group: failed to delete", "err", err)
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), "groups.errors.delete_failed"))
		err = utils.SSEHub.Redirect(c, "/dashboard")
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

	err = utils.SSEHub.Redirect(c, "/dashboard")
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// ViewersPage shows the current viewers and invite form
func (g *Group) ViewersPage(c echo.Context) error {
	return g.renderViewersTab(c, viewersTabViewers)
}

func (g *Group) ViewersAdminsPage(c echo.Context) error {
	return g.renderViewersTab(c, viewersTabAdmins)
}

func (g *Group) ViewersPendingPage(c echo.Context) error {
	return g.renderViewersTab(c, viewersTabPending)
}

func (g *Group) renderViewersTab(c echo.Context, tab string) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	data, err := g.viewersPageData(c, groupID, tab, c.QueryParams())
	if err != nil {
		slog.Error("group: failed to load viewers page", "group_id", groupID, "tab", tab, "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load viewers")
	}

	return utils.RenderPage(c, GroupViewersPage(data))
}

func (g *Group) patchViewersPage(c echo.Context, groupID, messageKey, errorKey string) error {
	if messageKey != "" {
		utils.Notify(c, "success", ctxi18n.T(c.Request().Context(), messageKey))
	}
	if errorKey != "" {
		utils.Notify(c, "error", ctxi18n.T(c.Request().Context(), errorKey))
	}
	tab, values := viewersTabAndQueryFromReferer(c)
	data, err := g.viewersPageData(c, groupID, tab, values)
	if err != nil {
		slog.Error("group: failed to load viewers patch data", "group_id", groupID, "tab", tab, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	html, err := utils.RenderHTMLForRequest(c, GroupViewersPage(data))
	if err != nil {
		slog.Error("group: failed to render viewers page", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	utils.SSEHub.PatchHTML(c, html)
	return c.NoContent(http.StatusOK)
}

// AddViewer adds an existing user as a group reader
func (g *Group) AddViewer(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	signals := addViewerSignals{}
	if err := datastar.ReadSignals(c.Request(), &signals); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	signals.FormData.Email = strings.ToLower(strings.TrimSpace(signals.FormData.Email))
	if errs := utils.ValidateWithLocale(c.Request().Context(), signals.FormData); errs != nil {
		utils.Notify(c, "error", errs["email"])
		return g.patchViewersPage(c, groupID, "", "")
	}
	emailAdress := signals.FormData.Email
	var err error

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return g.patchViewersPage(c, groupID, "", "groups.errors.group_not_found")
	}

	// If user exists and already has access, short-circuit
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), emailAdress)
	if err == nil {
		if group.AdminUserID == user.ID {
			return g.patchViewersPage(c, groupID, "groups.messages.already_admin", "")
		}
		count, err := db.Qry.IsGroupReader(c.Request().Context(), db.IsGroupReaderParams{
			UserID:  user.ID,
			GroupID: groupID,
		})
		if err == nil && count > 0 {
			return g.patchViewersPage(c, groupID, "groups.messages.already_viewer", "")
		}
	}

	// Create invite magic link
	token := utils.GenerateID("tok")
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = db.Qry.CreateMagicLink(c.Request().Context(), db.CreateMagicLinkParams{
		ID:        utils.GenerateID("mag"),
		Token:     token,
		Email:     emailAdress,
		Action:    "invite",
		GroupID:   sql.NullString{String: groupID, Valid: true},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("group: failed to create invite link", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.invite_failed")
	}

	err = email.Email().SendGroupInvitation(c.Request().Context(), emailAdress, group.Name, token, utils.Env().URL)
	if err != nil {
		slog.Error("group: failed to send invite email", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.send_failed")
	}

	utils.SSEHub.PatchSignals(c, map[string]any{"formState": "", "formData": map[string]any{"email": ""}})
	return g.patchViewersPage(c, groupID, "groups.messages.invite_sent", "")
}

// RemoveViewer removes a reader from the group
func (g *Group) RemoveViewer(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if !utils.IsValidID(userID, "usr") {
		return g.patchViewersPage(c, groupID, "", "groups.errors.invalid_user")
	}

	err := db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("group: failed to remove viewer", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.remove_failed")
	}

	return g.patchViewersPage(c, groupID, "groups.messages.viewer_removed", "")
}

// CancelInvite removes a pending invitation from the group.
func (g *Group) CancelInvite(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	inviteID := c.Param("inviteId")
	if !utils.IsValidID(inviteID, "mag") {
		return g.patchViewersPage(c, groupID, "", "groups.errors.invalid_invite")
	}

	err := db.Qry.DeleteGroupPendingInvite(c.Request().Context(), db.DeleteGroupPendingInviteParams{
		ID:      inviteID,
		GroupID: sql.NullString{String: groupID, Valid: true},
	})
	if err != nil {
		slog.Error("group: failed to cancel invite", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.invite_cancel_failed")
	}

	return g.patchViewersPage(c, groupID, "groups.messages.invite_cancelled", "")
}

func (g *Group) viewersPageData(c echo.Context, groupID, tab string, values url.Values) (ViewersPageData, error) {
	ctx := c.Request().Context()
	group, err := db.Qry.GetGroupByID(ctx, groupID)
	if err != nil {
		return ViewersPageData{}, err
	}

	admin, err := db.Qry.GetUserByID(ctx, group.AdminUserID)
	if err != nil {
		return ViewersPageData{}, err
	}

	admins := []db.User{admin}

	var queryable utils.Queryable = g.viewersModel
	switch tab {
	case viewersTabAdmins:
		queryable = g.adminsModel
	case viewersTabPending:
		queryable = g.pendingInvitesModel
	default:
		tab = viewersTabViewers
	}

	query := parseTableQueryFromValues(values, queryable)

	var viewers []db.User
	var invites []db.MagicLink
	var total int64

	switch tab {
	case viewersTabAdmins:
		filteredAdmins := filterAdminsByQuery(admins, query)
		total = int64(len(filteredAdmins))
		query = utils.ClampPage(query, total)
		admins = pageAdmins(filteredAdmins, query)
	case viewersTabPending:
		total, err = db.Qry.CountGroupPendingInvitesFiltered(ctx, db.CountGroupPendingInvitesFilteredParams{
			GroupID: sql.NullString{String: groupID, Valid: true},
			Search:  query.Search,
		})
		if err != nil {
			return ViewersPageData{}, err
		}
		query = utils.ClampPage(query, total)

		switch query.Sort {
		case "email":
			if query.Dir == "desc" {
				invites, err = db.Qry.ListGroupPendingInvitesByEmailDescFiltered(ctx, db.ListGroupPendingInvitesByEmailDescFilteredParams{
					GroupID: sql.NullString{String: groupID, Valid: true},
					Search:  query.Search,
					Limit:   int64(query.PageSize),
					Offset:  query.Offset(),
				})
			} else {
				invites, err = db.Qry.ListGroupPendingInvitesByEmailAscFiltered(ctx, db.ListGroupPendingInvitesByEmailAscFilteredParams{
					GroupID: sql.NullString{String: groupID, Valid: true},
					Search:  query.Search,
					Limit:   int64(query.PageSize),
					Offset:  query.Offset(),
				})
			}
		default:
			if query.Dir == "asc" {
				invites, err = db.Qry.ListGroupPendingInvitesByCreatedAscFiltered(ctx, db.ListGroupPendingInvitesByCreatedAscFilteredParams{
					GroupID: sql.NullString{String: groupID, Valid: true},
					Search:  query.Search,
					Limit:   int64(query.PageSize),
					Offset:  query.Offset(),
				})
			} else {
				invites, err = db.Qry.ListGroupPendingInvitesByCreatedDescFiltered(ctx, db.ListGroupPendingInvitesByCreatedDescFilteredParams{
					GroupID: sql.NullString{String: groupID, Valid: true},
					Search:  query.Search,
					Limit:   int64(query.PageSize),
					Offset:  query.Offset(),
				})
			}
		}
		if err != nil {
			return ViewersPageData{}, err
		}
	default:
		total, err = db.Qry.CountGroupReadersFiltered(ctx, db.CountGroupReadersFilteredParams{
			GroupID: groupID,
			Search:  query.Search,
		})
		if err != nil {
			return ViewersPageData{}, err
		}
		query = utils.ClampPage(query, total)

		if query.Dir == "desc" {
			viewers, err = db.Qry.ListGroupReadersByEmailDescFiltered(ctx, db.ListGroupReadersByEmailDescFilteredParams{
				GroupID: groupID,
				Search:  query.Search,
				Limit:   int64(query.PageSize),
				Offset:  query.Offset(),
			})
		} else {
			viewers, err = db.Qry.ListGroupReadersByEmailAscFiltered(ctx, db.ListGroupReadersByEmailAscFilteredParams{
				GroupID: groupID,
				Search:  query.Search,
				Limit:   int64(query.PageSize),
				Offset:  query.Offset(),
			})
		}
		if err != nil {
			return ViewersPageData{}, err
		}
	}

	return ViewersPageData{
		Title: ctxi18n.T(ctx, "groups.access_page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"},
			{Label: group.Name, Href: "/groups/" + group.ID},
			{Label: ctxi18n.T(ctx, "groups.access"), Href: "/groups/" + group.ID + "/access/viewers"},
			{Label: accessTabLabel(ctx, tab)},
		},
		UserEmail:    getUserEmail(c),
		Group:        group,
		Admins:       admins,
		Viewers:      viewers,
		Invites:      invites,
		IsAdmin:      middleware.IsAdmin(c),
		Query:        query,
		Pager:        utils.BuildTablePagination(total, query),
		GroupID:      groupID,
		Tab:          tab,
		AdminsTable:  utils.ViewersAdminsTableLayout(),
		PendingTable: utils.ViewersPendingTableLayout(),
		ViewersTable: utils.ViewersTableLayout(),
	}, nil
}

func accessTabLabel(ctx context.Context, tab string) string {
	switch tab {
	case viewersTabAdmins:
		return ctxi18n.T(ctx, "groups.admins")
	case viewersTabPending:
		return ctxi18n.T(ctx, "groups.pending_invitations")
	default:
		return ctxi18n.T(ctx, "groups.viewers")
	}
}

func accessTabPath(groupID, tab string) string {
	switch tab {
	case viewersTabAdmins:
		return "/groups/" + groupID + "/access/admins"
	case viewersTabPending:
		return "/groups/" + groupID + "/access/pending"
	default:
		return "/groups/" + groupID + "/access/viewers"
	}
}

func viewersTabAndQueryFromReferer(c echo.Context) (string, url.Values) {
	referer := c.Request().Referer()
	if referer == "" {
		return viewersTabViewers, url.Values{}
	}

	u, err := url.Parse(referer)
	if err != nil {
		return viewersTabViewers, url.Values{}
	}

	return viewersTabFromPath(u.Path), u.Query()
}

func viewersTabFromPath(path string) string {
	switch {
	case strings.HasSuffix(path, "/access/admins"):
		return viewersTabAdmins
	case strings.HasSuffix(path, "/access/pending"):
		return viewersTabPending
	case strings.HasSuffix(path, "/access/viewers"):
		return viewersTabViewers
	default:
		return viewersTabViewers
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

func filterAdminsByQuery(admins []db.User, query utils.TableQuery) []db.User {
	if query.Search == "" {
		return admins
	}

	filtered := make([]db.User, 0, len(admins))
	needle := strings.ToLower(query.Search)
	for _, admin := range admins {
		if strings.Contains(strings.ToLower(admin.Email), needle) {
			filtered = append(filtered, admin)
		}
	}

	return filtered
}

func pageAdmins(admins []db.User, query utils.TableQuery) []db.User {
	if len(admins) == 0 {
		return admins
	}

	start := int(query.Offset())
	if start >= len(admins) {
		return []db.User{}
	}

	end := start + query.PageSize
	if end > len(admins) {
		end = len(admins)
	}

	return admins[start:end]
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

	return GroupPageData{
		Title:       "Bandcash - " + group.Name,
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(ctx, "groups.title"), Href: "/dashboard"}, {Label: group.Name, Href: "/groups/" + groupID}, {Label: ctxi18n.T(ctx, "nav.overview")}},
		UserEmail:   getUserEmail(c),
		Group:       group,
		Admin:       admin,
		Income:      group.TotalEventAmount,
		Payouts:     group.TotalPayoutAmount,
		Expenses:    group.TotalExpenseAmount,
		Leftover:    group.TotalLeftover,
		IsAdmin:     middleware.IsAdmin(c),
	}, nil
}
