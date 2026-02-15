package utils

import (
	"bytes"
	"html/template"
)

func RenderBlock(tmpl *template.Template, blockName string, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, blockName, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
