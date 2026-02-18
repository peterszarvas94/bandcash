package group

import (
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
