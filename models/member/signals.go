package member

func memberIndexSignals(csrfToken string) map[string]any {
	return map[string]any{
		"csrf":      csrfToken,
		"mode":      "",
		"formState": "",
		"editingId": 0,
		"formData":  map[string]any{"name": "", "description": ""},
	}
}

func memberShowSignals(data MemberData, csrfToken string) map[string]any {
	return map[string]any{
		"csrf":      csrfToken,
		"mode":      "single",
		"formState": "",
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
