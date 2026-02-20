package group

import (
	"database/sql"
	"log/slog"
	"net/http"
	"net/url"
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
	emailService *email.Service
	appBaseURL   string
}

type createGroupSignals struct {
	FormData struct {
		Name string `json:"name"`
	} `json:"formData"`
}

type addViewerSignals struct {
	FormData struct {
		Email string `json:"email"`
	} `json:"formData"`
}

func New() *Group {
	cfg := utils.Env()
	return &Group{
		emailService: email.NewFromEnv(),
		appBaseURL:   cfg.URL,
	}
}

// NewGroupPage shows the form to create a new group
func (g *Group) NewGroupPage(c echo.Context) error {
	utils.EnsureClientID(c)
	userEmail := getUserEmail(c)
	data := NewGroupPageData{
		Title:       ctxi18n.T(c.Request().Context(), "groups.new"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/dashboard"}, {Label: ctxi18n.T(c.Request().Context(), "groups.new")}},
		UserEmail:   userEmail,
	}
	return utils.RenderComponent(c, GroupNewPage(data))
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

	name := signals.FormData.Name
	if name == "" {
		return c.NoContent(http.StatusBadRequest)
	}

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

	// Redirect to events
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

	adminGroups, err := db.Qry.ListGroupsByAdmin(c.Request().Context(), userID)
	if err != nil {
		slog.Error("group: failed to load admin groups", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load groups")
	}

	readerGroups, err := db.Qry.ListGroupsByReader(c.Request().Context(), userID)
	if err != nil {
		slog.Error("group: failed to load reader groups", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load groups")
	}

	// Remove any reader groups where user is admin
	adminMap := make(map[string]bool, len(adminGroups))
	for _, group := range adminGroups {
		adminMap[group.ID] = true
	}
	filteredReaders := make([]db.Group, 0, len(readerGroups))
	for _, group := range readerGroups {
		if adminMap[group.ID] {
			continue
		}
		filteredReaders = append(filteredReaders, group)
	}

	data := GroupsPageData{
		Title:        ctxi18n.T(c.Request().Context(), "groups.title"),
		Breadcrumbs:  []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title")}},
		UserEmail:    userEmail,
		AdminGroups:  adminGroups,
		ReaderGroups: filteredReaders,
		MessageKey:   c.QueryParam("msg"),
		ErrorKey:     c.QueryParam("error"),
	}
	return utils.RenderComponent(c, GroupsPage(data))
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
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.group_not_found"))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if group.AdminUserID == userID {
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.admin_cannot_leave"))
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
		slog.Warn("group: failed to leave", "err", err)
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.leave_failed"))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	err = utils.SSEHub.Redirect(c, "/dashboard?msg="+url.QueryEscape("groups.messages.left"))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// DeleteGroup removes the group and all data (admin only).
func (g *Group) DeleteGroup(c echo.Context) error {
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
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.group_not_found"))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}
	if group.AdminUserID != userID {
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.admin_required"))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	if err := db.Qry.DeleteGroup(c.Request().Context(), groupID); err != nil {
		slog.Error("group: failed to delete", "err", err)
		err = utils.SSEHub.Redirect(c, "/dashboard?error="+url.QueryEscape("groups.errors.delete_failed"))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		return c.NoContent(http.StatusOK)
	}

	err = utils.SSEHub.Redirect(c, "/dashboard?msg="+url.QueryEscape("groups.messages.deleted"))
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}
	return c.NoContent(http.StatusOK)
}

// ViewersPage shows the current viewers and invite form
func (g *Group) ViewersPage(c echo.Context) error {
	utils.EnsureClientID(c)
	groupID := middleware.GetGroupID(c)
	userEmail := getUserEmail(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return c.String(http.StatusNotFound, "Group not found")
	}

	viewers, err := db.Qry.GetGroupReaders(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group: failed to load viewers", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load viewers")
	}

	data := ViewersPageData{
		Title:       ctxi18n.T(c.Request().Context(), "groups.viewers"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/dashboard"}, {Label: group.Name, Href: "/groups/" + group.ID}, {Label: ctxi18n.T(c.Request().Context(), "groups.viewers")}},
		UserEmail:   userEmail,
		Group:       group,
		Viewers:     viewers,
		MessageKey:  c.QueryParam("msg"),
		ErrorKey:    c.QueryParam("error"),
	}

	return utils.RenderComponent(c, GroupViewersPage(data))
}

func (g *Group) patchViewersPage(c echo.Context, groupID, messageKey, errorKey string) error {
	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group: failed to load group for viewers patch", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	viewers, err := db.Qry.GetGroupReaders(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group: failed to load viewers for patch", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := ViewersPageData{
		Title:       ctxi18n.T(c.Request().Context(), "groups.viewers"),
		Breadcrumbs: []utils.Crumb{{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/dashboard"}, {Label: group.Name, Href: "/groups/" + group.ID}, {Label: ctxi18n.T(c.Request().Context(), "groups.viewers")}},
		UserEmail:   getUserEmail(c),
		Group:       group,
		Viewers:     viewers,
		MessageKey:  messageKey,
		ErrorKey:    errorKey,
	}

	html, err := utils.RenderComponentString(c.Request().Context(), GroupViewersPage(data))
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
	email := signals.FormData.Email
	var err error
	if email == "" {
		return g.patchViewersPage(c, groupID, "", "groups.errors.email_required")
	}

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return g.patchViewersPage(c, groupID, "", "groups.errors.group_not_found")
	}

	// If user exists and already has access, short-circuit
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), email)
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
		Email:     email,
		Action:    "invite",
		GroupID:   sql.NullString{String: groupID, Valid: true},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		slog.Error("group: failed to create invite link", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.invite_failed")
	}

	err = g.emailService.SendGroupInvitation(email, group.Name, token, g.appBaseURL)
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
	if userID == "" {
		return g.patchViewersPage(c, groupID, "", "groups.errors.invalid_user")
	}

	err := db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Warn("group: failed to remove viewer", "err", err)
		return g.patchViewersPage(c, groupID, "", "groups.errors.remove_failed")
	}

	return g.patchViewersPage(c, groupID, "groups.messages.viewer_removed", "")
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
