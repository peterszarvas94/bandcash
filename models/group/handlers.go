package group

import (
	"database/sql"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/email"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Group struct {
	emailService *email.Service
}

func New() *Group {
	return &Group{
		emailService: email.NewFromEnv(),
	}
}

// NewGroupPage shows the form to create a new group
func (g *Group) NewGroupPage(c echo.Context) error {
	return c.HTML(http.StatusOK, newGroupPageHTML)
}

// CreateGroup handles group creation
func (g *Group) CreateGroup(c echo.Context) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Redirect(http.StatusFound, "/auth/login")
	}

	name := c.FormValue("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "Group name required")
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
	return c.Redirect(http.StatusFound, "/groups/"+group.ID+"/events")
}

// GroupsPage lists groups the user can access
func (g *Group) GroupsPage(c echo.Context) error {
	userID := middleware.GetUserID(c)
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

	return c.HTML(http.StatusOK, renderGroupsPage(adminGroups, filteredReaders))
}

// ViewersPage shows the current viewers and invite form
func (g *Group) ViewersPage(c echo.Context) error {
	groupID := middleware.GetGroupID(c)

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return c.String(http.StatusNotFound, "Group not found")
	}

	viewers, err := db.Qry.GetGroupReaders(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("group: failed to load viewers", "err", err)
		return c.String(http.StatusInternalServerError, "Failed to load viewers")
	}

	msg := c.QueryParam("msg")
	errMsg := c.QueryParam("error")

	return c.HTML(http.StatusOK, renderViewersPage(group, viewers, msg, errMsg))
}

// AddViewer adds an existing user as a group reader
func (g *Group) AddViewer(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	email := c.FormValue("email")
	if email == "" {
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Email%20required")
	}

	group, err := db.Qry.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Group%20not%20found")
	}

	// If user exists and already has access, short-circuit
	user, err := db.Qry.GetUserByEmail(c.Request().Context(), email)
	if err == nil {
		if group.AdminUserID == user.ID {
			return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?msg=User%20is%20already%20admin")
		}
		count, err := db.Qry.IsGroupReader(c.Request().Context(), db.IsGroupReaderParams{
			UserID:  user.ID,
			GroupID: groupID,
		})
		if err == nil && count > 0 {
			return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?msg=User%20already%20viewer")
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
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Failed%20to%20create%20invite")
	}

	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	err = g.emailService.SendGroupInvitation(email, group.Name, token, baseURL)
	if err != nil {
		slog.Error("group: failed to send invite email", "err", err)
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Failed%20to%20send%20email")
	}

	return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?msg=Invite%20sent")
}

// RemoveViewer removes a reader from the group
func (g *Group) RemoveViewer(c echo.Context) error {
	groupID := middleware.GetGroupID(c)
	userID := c.Param("userId")
	if userID == "" {
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Invalid%20user")
	}

	err := db.Qry.RemoveGroupReader(c.Request().Context(), db.RemoveGroupReaderParams{
		UserID:  userID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Warn("group: failed to remove viewer", "err", err)
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=Failed%20to%20remove")
	}

	return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?msg=Viewer%20removed")
}

var newGroupPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>Create Group - BandCash</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Create Your Band Group</h1>
        <p>Welcome! Let's set up your band's money management system.</p>
        <form method="POST" action="/groups">
            <div class="field">
                <label>Group Name</label>
                <input type="text" name="name" required placeholder="My Awesome Band">
            </div>
            <button type="submit" class="btn btn-primary">Create Group</button>
        </form>
        <p><a href="/auth/logout">Logout</a></p>
    </div>
</body>
</html>`

func renderGroupsPage(adminGroups, readerGroups []db.Group) string {
	adminRows := ""
	if len(adminGroups) == 0 {
		adminRows = `<tr><td colspan="2">No groups yet.</td></tr>`
	} else {
		for _, group := range adminGroups {
			adminRows += fmt.Sprintf(
				`<tr><td>%s</td><td><a class="btn" href="/groups/%s/events">Open</a></td></tr>`,
				html.EscapeString(group.Name),
				html.EscapeString(group.ID),
			)
		}
	}

	readerRows := ""
	if len(readerGroups) == 0 {
		readerRows = `<tr><td colspan="2">No viewer access.</td></tr>`
	} else {
		for _, group := range readerGroups {
			readerRows += fmt.Sprintf(
				`<tr><td>%s</td><td><a class="btn" href="/groups/%s/events">Open</a></td></tr>`,
				html.EscapeString(group.Name),
				html.EscapeString(group.ID),
			)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Your Groups - BandCash</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Your Groups</h1>
        <p><a class="btn btn-primary" href="/groups/new">Create New Group</a></p>

        <h2>Admin Groups</h2>
        <table class="table">
            <thead><tr><th>Name</th><th></th></tr></thead>
            <tbody>%s</tbody>
        </table>

        <h2>Viewer Groups</h2>
        <table class="table">
            <thead><tr><th>Name</th><th></th></tr></thead>
            <tbody>%s</tbody>
        </table>

        <p><a href="/auth/logout">Logout</a></p>
    </div>
</body>
</html>`,
		adminRows,
		readerRows,
	)
}

func renderViewersPage(group db.Group, viewers []db.User, msg, errMsg string) string {
	messageHTML := ""
	if msg != "" {
		messageHTML = `<p class="notice">` + html.EscapeString(msg) + `</p>`
	}
	if errMsg != "" {
		messageHTML = `<p class="error">` + html.EscapeString(errMsg) + `</p>`
	}

	rows := ""
	if len(viewers) == 0 {
		rows = `<tr><td colspan="2">No viewers yet.</td></tr>`
	} else {
		for _, viewer := range viewers {
			rows += fmt.Sprintf(
				`<tr><td>%s</td><td>
                <form method="POST" action="/groups/%s/viewers/%s/remove" style="display:inline">
                    <button type="submit" class="btn btn-danger">Remove</button>
                </form>
                </td></tr>`,
				html.EscapeString(viewer.Email),
				html.EscapeString(group.ID),
				html.EscapeString(viewer.ID),
			)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Viewers - %s</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <h1>Viewers for %s</h1>
        %s

        <h2>Invite Viewer</h2>
        <form method="POST" action="/groups/%s/viewers">
            <div class="field">
                <label>Email</label>
                <input type="email" name="email" required placeholder="viewer@email.com">
            </div>
            <button type="submit" class="btn btn-primary">Add Viewer</button>
        </form>

        <h2>Current Viewers</h2>
        <table class="table">
            <thead>
                <tr><th>Email</th><th>Actions</th></tr>
            </thead>
            <tbody>%s</tbody>
        </table>

        <p><a href="/groups/%s/events">Back to Events</a></p>
    </div>
</body>
</html>`,
		html.EscapeString(group.Name),
		html.EscapeString(group.Name),
		messageHTML,
		html.EscapeString(group.ID),
		rows,
		html.EscapeString(group.ID),
	)
}
