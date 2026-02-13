package entry

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/sse"
)

type SSERenderer struct {
	indexTmpl *template.Template
	showTmpl  *template.Template
	editTmpl  *template.Template
	entries   *Entries
	routes    []entryViewRoute
}

type entryViewRoute struct {
	pattern string
	render  func(c echo.Context, params map[string]string) (string, error)
}

func NewSSERenderer() *SSERenderer {
	editTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/edit.html",
	))
	showTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/show.html",
	))
	indexTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/index.html",
	))

	r := &SSERenderer{
		indexTmpl: indexTmpl,
		showTmpl:  showTmpl,
		editTmpl:  editTmpl,
		entries:   New(),
	}

	r.routes = []entryViewRoute{
		{pattern: "/entry", render: r.renderIndex},
		{pattern: "/entry/:id", render: r.renderShow},
		{pattern: "/entry/:id/edit", render: r.renderEdit},
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

	return sse.RenderTemplate(r.indexTmpl, data)
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

	return sse.RenderTemplate(r.showTmpl, data)
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

	return sse.RenderTemplate(r.editTmpl, data)
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
