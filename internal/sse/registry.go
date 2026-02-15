package sse

import (
	"bytes"
	"errors"
	"html/template"

	"github.com/labstack/echo/v4"
)

// ErrViewNotFound signals a renderer doesn't handle the view.
var ErrViewNotFound = errors.New("view not found")

type Registry struct {
	renderers []ViewRenderer
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Add(renderer ViewRenderer) {
	r.renderers = append(r.renderers, renderer)
}

func (r *Registry) Render(c echo.Context, view string) (string, error) {
	for _, renderer := range r.renderers {
		html, err := renderer(c, view)
		if err == nil {
			return html, nil
		}
		if errors.Is(err, ErrViewNotFound) {
			continue
		}
		return "", err
	}

	return "", echo.NewHTTPError(404, "View not found")
}

func RenderTemplate(tmpl *template.Template, blockName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, blockName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
