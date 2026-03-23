package member

import "bandcash/internal/utils"

func memberIndexSignals(csrfToken string, query map[string]any) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"tableQuery":     query,
		"mode":           "table",
		"formState":      "",
		"eventFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"name": "", "description": ""},
		"errors":         map[string]any{"name": "", "description": ""},
	}
}

func memberShowSignals(data MemberData, csrfToken string) map[string]any {
	return map[string]any{
		"csrf":           csrfToken,
		"mode":           "single",
		"tableQuery":     utils.TableQuerySignals(data.Query),
		"formState":      "",
		"eventFormState": "",
		"summaryMode":    "all",
		"formData": map[string]any{
			"name":        data.Member.Name,
			"description": data.Member.Description,
		},
		"errors": map[string]any{
			"name":        "",
			"description": "",
		},
	}
}
