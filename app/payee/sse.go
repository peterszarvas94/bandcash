package payee

import (
	"html/template"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/sse"
	"bandcash/internal/view"
)

type SSERenderer struct {
	tmpl   *template.Template
	payees *Payees
	routes []payeeViewRoute
}

type payeeViewRoute struct {
	pattern   string
	blockName string
	render    func(c echo.Context, params map[string]string) (string, error)
}

func NewSSERenderer(tmpl *template.Template) *SSERenderer {
	r := &SSERenderer{
		tmpl:   tmpl,
		payees: New(),
	}

	r.routes = []payeeViewRoute{
		{pattern: "/payee", blockName: "payee-index", render: r.renderIndex},
		{pattern: "/payee/new", blockName: "payee-new", render: r.renderNew},
		{pattern: "/payee/:id", blockName: "payee-show", render: r.renderShow},
		{pattern: "/payee/:id/edit", blockName: "payee-edit", render: r.renderEdit},
	}

	return r
}

func (r *SSERenderer) Render(c echo.Context, view string) (string, error) {
	view = strings.TrimSuffix(view, "/")

	for _, route := range r.routes {
		params, ok := matchPayeeView(route.pattern, view)
		if !ok {
			continue
		}

		return route.render(c, params)
	}

	return "", sse.ErrViewNotFound
}

func (r *SSERenderer) renderIndex(c echo.Context, _ map[string]string) (string, error) {
	data, err := r.payees.GetIndexData(c.Request().Context())
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "payee-index", data)
}

func (r *SSERenderer) renderNew(c echo.Context, _ map[string]string) (string, error) {
	data := map[string]any{
		"Title":       "New Payee",
		"Breadcrumbs": []view.Crumb{{Label: "Payees", Href: "/payee"}, {Label: "New"}},
	}
	return sse.RenderTemplate(r.tmpl, "payee-new", data)
}

func (r *SSERenderer) renderShow(c echo.Context, params map[string]string) (string, error) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		return "", echo.NewHTTPError(400, "Invalid payee ID")
	}

	data, err := r.payees.GetShowData(c.Request().Context(), id)
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "payee-show", data)
}

func (r *SSERenderer) renderEdit(c echo.Context, params map[string]string) (string, error) {
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		return "", echo.NewHTTPError(400, "Invalid payee ID")
	}

	data, err := r.payees.GetEditData(c.Request().Context(), id)
	if err != nil {
		return "", err
	}

	return sse.RenderTemplate(r.tmpl, "payee-edit", data)
}

func matchPayeeView(pattern, view string) (map[string]string, bool) {
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
