package dev

import "bandcash/internal/utils"

type DevPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	LinkSelector    string
	IsAuthenticated bool
	UserEmail       string
}

func DevPageSignals() map[string]any {
	return map[string]any{
		"errors":           map[string]any{"name": ""},
		"formData":         map[string]any{"name": ""},
		"_fetching":        false,
		"activeSpinner":    "",
		"_notifyInline":    false,
		"_notifyTest":      false,
		"switchOff":        false,
		"switchOn":         true,
		"switchDisabled":   false,
		"switchDisabledOn": true,
		"radioDemo":        "1",
	}
}
