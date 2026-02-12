package payee

import (
	"html/template"
	"strings"

	"github.com/labstack/echo/v4"

	"bandcash/internal/sse"
)

type SSERenderer struct {
	indexTmpl *template.Template
	payees    *Payees
}

func NewSSERenderer() *SSERenderer {
	indexTmpl := template.Must(template.ParseFiles(
		"web/templates/breadcrumbs.html",
		"app/payee/templates/index.html",
	))

	return &SSERenderer{
		indexTmpl: indexTmpl,
		payees:    New(),
	}
}

func (r *SSERenderer) Render(c echo.Context, view string) (string, error) {
	view = strings.TrimSuffix(view, "/")

	if view == "/payee" {
		data, err := r.payees.GetIndexData(c.Request().Context())
		if err != nil {
			return "", err
		}

		return sse.RenderTemplate(r.indexTmpl, data)
	}

	return "", sse.ErrViewNotFound
}
