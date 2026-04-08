package member

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/db"
	"bandcash/internal/middleware"
	"bandcash/internal/utils"
)

func NewMemberPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	group, err := db.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.new_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := NewMemberPageData{
		Title: ctxi18n.T(c.Request().Context(), "members.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "members.title"), Href: "/groups/" + groupID + "/members"},
			{Label: ctxi18n.T(c.Request().Context(), "members.add")},
		},
		GroupID: groupID,
		Signals: map[string]any{
			"formData": map[string]any{"name": "", "description": ""},
			"errors":   map[string]any{"name": "", "description": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    middleware.IsSuperadmin(c),
	}
	return utils.RenderPage(c, MemberNewPage(data))
}

func EditMemberPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Info("member.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	group, err := db.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("member.edit_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	member, err := db.GetMember(c.Request().Context(), db.GetMemberParams{
		ID:      id,
		GroupID: groupID,
	})
	if err != nil {
		slog.Error("member.edit_page: failed to get member", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := EditMemberPageData{
		Title: ctxi18n.T(c.Request().Context(), "members.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "members.title"), Href: "/groups/" + groupID + "/members"},
			{Label: member.Name, Href: "/groups/" + groupID + "/members/" + id},
			{Label: ctxi18n.T(c.Request().Context(), "members.edit")},
		},
		GroupID: groupID,
		Member:  &member,
		Signals: map[string]any{
			"formData": map[string]any{"name": member.Name, "description": member.Description},
			"errors":   map[string]any{"name": "", "description": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    middleware.IsSuperadmin(c),
	}
	return utils.RenderPage(c, MemberEditPage(data))
}

func Index(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: TableQuerySpec()})

	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("member.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	data.IsAdmin = middleware.IsAdmin(c)
	data.Signals = memberIndexSignals(utils.TableQuerySignals(data.Query))
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	slog.Debug("member.index", "member_count", len(data.Members))
	return utils.RenderPage(c, MemberIndex(data))
}

func Show(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := middleware.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: MemberEventsTableQuerySpec()})

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixMember) {
		slog.Info("member.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("member.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyMemberShowTableByRole(&data, middleware.IsAdmin(c))
	data.Signals = memberShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = middleware.IsSuperadmin(c)

	return utils.RenderPage(c, MemberShow(data))
}
