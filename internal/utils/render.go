package utils

import (
	"bytes"
	"html/template"
	"io"
)

func RenderBlock(tmpl *template.Template, blockName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, blockName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RenderTemplate(w io.Writer, tmpl *template.Template, name string, data any) error {
	return tmpl.ExecuteTemplate(w, name, data)
}
