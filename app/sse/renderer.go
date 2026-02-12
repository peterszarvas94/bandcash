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
	entryTmpl *template.Template
	entries   *entry.Entries
}

func NewRenderer() *Renderer {
	entryTmpl := template.Must(template.ParseFiles(
		"app/entry/templates/show.html",
	))

	return &Renderer{
		entryTmpl: entryTmpl,
		entries:   entry.New(),
	}
}

func (r *Renderer) Render(c echo.Context, view string) (string, error) {
	view = strings.TrimSuffix(view, "/")

	if strings.HasPrefix(view, "/entry/") {
		idStr := strings.TrimPrefix(view, "/entry/")
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

	return "", echo.NewHTTPError(404, "View not found")
}
