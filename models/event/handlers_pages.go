package event

import (
	"log/slog"
	"net/http"

	ctxi18n "github.com/invopop/ctxi18n/i18n"
	"github.com/labstack/echo/v4"

	"bandcash/internal/utils"
	groupstore "bandcash/models/group/data"
)

func IndexPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := utils.ParseTableQuery(c, staticTableQueryable{spec: TableQuerySpec()})

	data, err := GetIndexData(c.Request().Context(), groupID, query)
	if err != nil {
		slog.Error("event.list: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventIndexTableByRole(&data, utils.IsAdmin(c))
	data.Signals = eventIndexSignals(data.Query)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	slog.Debug("event.index", "event_count", len(data.Events))
	return utils.RenderPage(c, EventIndexPage(data))
}

func ShowPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := parseParticipantTableQuery(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.show: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.show: failed to get data", "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	applyEventShowTableByRole(&data, utils.IsAdmin(c))
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, EventShowPage(data))
}

func NewEventPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)

	group, err := groupstore.GetGroupByID(c.Request().Context(), groupID)
	if err != nil {
		slog.Error("event.new_page: failed to get group", "group_id", groupID, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	data := NewEventPageData{
		Title: ctxi18n.T(c.Request().Context(), "events.page_title"),
		Breadcrumbs: []utils.Crumb{
			{Label: ctxi18n.T(c.Request().Context(), "groups.title"), Href: "/groups"},
			{Label: group.Name, Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "events.title"), Href: "/groups/" + groupID + "/events"},
			{Label: ctxi18n.T(c.Request().Context(), "events.add")},
		},
		GroupID: groupID,
		Signals: map[string]any{
			"formData": map[string]any{"title": "", "date": "", "time": "", "place": "", "description": "", "amount": 0, "paid": false, "paidAt": ""},
			"errors":   map[string]any{"title": "", "date": "", "time": "", "place": "", "description": "", "amount": ""},
		},
		IsAuthenticated: true,
		IsSuperAdmin:    utils.IsSuperadmin(c),
	}
	return utils.RenderPage(c, EventNewPage(data))
}

func EditEventPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := parseParticipantTableQuery(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.edit_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.edit_page: failed to get data", "group_id", groupID, "event_id", id, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	applyEventShowTableByRole(&data, utils.IsAdmin(c))
	if len(data.Breadcrumbs) > 0 {
		data.Breadcrumbs[len(data.Breadcrumbs)-1].Href = "/groups/" + groupID + "/events/" + id
	}
	data.Breadcrumbs = append(data.Breadcrumbs, utils.Crumb{Label: ctxi18n.T(c.Request().Context(), "events.edit")})
	data.EditorMode = "edit_members"
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, EventEditPage(data))
}

func EditEventDetailsPage(c echo.Context) error {
	utils.EnsureTabID(c)
	groupID := utils.GetGroupID(c)
	query := parseParticipantTableQuery(c)

	id := c.Param("id")
	if !utils.IsValidID(id, utils.PrefixEvent) {
		slog.Info("event.edit_details_page: invalid id")
		return c.NoContent(http.StatusBadRequest)
	}

	data, err := GetShowData(c.Request().Context(), groupID, id, query)
	if err != nil {
		slog.Error("event.edit_details_page: failed to get data", "group_id", groupID, "event_id", id, "err", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	applyEventShowTableByRole(&data, utils.IsAdmin(c))
	if len(data.Breadcrumbs) > 0 {
		data.Breadcrumbs[len(data.Breadcrumbs)-1].Href = "/groups/" + groupID + "/events/" + id
	}
	data.Breadcrumbs = append(data.Breadcrumbs, utils.Crumb{Label: ctxi18n.T(c.Request().Context(), "events.edit")})
	data.EditorMode = "edit_details"
	data.Signals = eventShowSignals(data)
	data.IsAuthenticated = true
	data.IsSuperAdmin = utils.IsSuperadmin(c)

	return utils.RenderPage(c, EventEditPage(data))
}
