package utils

import "github.com/a-h/templ"

func JSONString(value any) string {
	result, err := templ.JSONString(value)
	if err != nil {
		return "null"
	}
	return result
}
