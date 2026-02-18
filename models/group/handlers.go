package group

import (
	"fmt"
	"html"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

type Group struct{}

func New() *Group {
	return &Group{}
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

	user, err := db.Qry.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=User%20not%20found")
	}

	_, err = db.Qry.CreateGroupReader(c.Request().Context(), db.CreateGroupReaderParams{
		ID:      utils.GenerateID("grd"),
		UserID:  user.ID,
		GroupID: groupID,
	})
	if err != nil {
		slog.Warn("group: failed to add viewer", "err", err)
		return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?error=User%20already%20viewer")
	}

	return c.Redirect(http.StatusFound, "/groups/"+groupID+"/viewers?msg=Viewer%20added")
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
