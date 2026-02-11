package home

import (
	"html/template"

	"github.com/labstack/echo/v4"
)

type HomeData struct {
	Title string
}

type Home struct {
	tmpl *template.Template
}

// Data returns data for rendering.
func (h *Home) Data() any {
	return HomeData{
		Title: "Bandcash",
	}
}

// Register registers home routes.
func Register(e *echo.Echo) {
	h := &Home{}

	// Parse shared head + home template
	h.tmpl = template.Must(template.ParseFiles(
		"web/templates/head.html",
		"app/home/templates/index.html",
	))

	e.GET("/", h.Index)
}
