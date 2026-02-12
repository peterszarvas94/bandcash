package sse

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/app/entry"
)

type Renderer struct {
	indexTmpl *template.Template
	entryTmpl *template.Template
	entries   *entry.Entries
}

func NewRenderer() *Renderer {
	entryTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/show.html",
	))
	indexTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/entry/templates/index.html",
	))

	return &Renderer{
		indexTmpl: indexTmpl,
		entryTmpl: entryTmpl,
		entries:   entry.New(),
	}
}

func (r *Renderer) Render(c echo.Context, view string) (string, error) {
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

		var buf bytes.Buffer
		if err := r.entryTmpl.ExecuteTemplate(&buf, "app", data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	if view == "/entry" {
		data, err := r.entries.GetIndexData(c.Request().Context())
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := r.indexTmpl.ExecuteTemplate(&buf, "app", data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	return "", echo.NewHTTPError(404, "View not found")
}
