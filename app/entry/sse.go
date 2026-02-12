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
	entries   *Entries
}

func NewSSERenderer() *SSERenderer {
	showTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/show.html",
	))
	indexTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/index.html",
	))

	return &SSERenderer{
		indexTmpl: indexTmpl,
		showTmpl:  showTmpl,
		entries:   New(),
	}
}

func (r *SSERenderer) Render(c echo.Context, view string) (string, error) {
	view = strings.TrimSuffix(view, "/")

	if idStr, ok := strings.CutPrefix(view, "/entry/"); ok {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return "", echo.NewHTTPError(400, "Invalid entry ID")
		}

		data, err := r.entries.GetShowData(c.Request().Context(), id)
		if err != nil {
			return "", err
		}

		return sse.RenderTemplate(r.showTmpl, data)
	}

	if view == "/entry" {
		data, err := r.entries.GetIndexData(c.Request().Context())
		if err != nil {
			return "", err
		}

		return sse.RenderTemplate(r.indexTmpl, data)
	}

	return "", sse.ErrViewNotFound
}
