package entry

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/sse"
	"bandcash/internal/view"
)

type SSERenderer struct {
	tmpl    *template.Template
	entries *Entries
	routes  []entryViewRoute
}

type entryViewRoute struct {
	pattern   string
	blockName string
	render    func(c echo.Context, params map[string]string) (string, error)
}

func NewSSERenderer(tmpl *template.Template) *SSERenderer {
	r := &SSERenderer{
		tmpl:    tmpl,
		entries: New(),
	}

	r.routes = []entryViewRoute{
		{pattern: "/entry", blockName: "entry-index", render: r.renderIndex},
		{pattern: "/entry/new", blockName: "entry-new", render: r.renderNew},
		{pattern: "/entry/:id", blockName: "entry-show", render: r.renderShow},
		{pattern: "/entry/:id/edit", blockName: "entry-edit", render: r.renderEdit},
	}

	return r
}

func (r *SSERenderer) Render(c echo.Context, view string) (string, error) {
	view = strings.TrimSuffix(view, "/")

	for _, route := range r.routes {
		params, ok := matchEntryView(route.pattern, view)
		if !ok {
			continue
		}

		return route.render(c, params)
	}

	return "", sse.ErrViewNotFound
}

func (r *SSERenderer) renderIndex(c echo.Context, _ map[string]string) (string, error) {
	data, err := r.entries.GetIndexData(c.Request().Context())
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "entry-index", data)
}

func (r *SSERenderer) renderNew(c echo.Context, _ map[string]string) (string, error) {
	data := map[string]any{
		"Title":       "New Entry",
		"Breadcrumbs": []view.Crumb{{Label: "Entries", Href: "/entry"}, {Label: "New"}},
	}
	return sse.RenderTemplate(r.tmpl, "entry-new", data)
}

func (r *SSERenderer) renderShow(c echo.Context, params map[string]string) (string, error) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		return "", echo.NewHTTPError(400, "Invalid entry ID")
	}

	data, err := r.entries.GetShowData(c.Request().Context(), id)
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "entry-show", data)
}

func (r *SSERenderer) renderEdit(c echo.Context, params map[string]string) (string, error) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		return "", echo.NewHTTPError(400, "Invalid entry ID")
	}

	data, err := r.entries.GetEditData(c.Request().Context(), id)
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "entry-edit", data)
}

func matchEntryView(pattern, view string) (map[string]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	viewParts := strings.Split(strings.Trim(view, "/"), "/")
	if len(patternParts) != len(viewParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := range patternParts {
		if name, ok := strings.CutPrefix(patternParts[i], ":"); ok {
			params[name] = viewParts[i]
			continue
		}
		if patternParts[i] != viewParts[i] {
			return nil, false
		}
	}

	return params, true
}
