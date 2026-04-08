package member

import "bandcash/internal/utils"

type memberParams struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type memberTableParams struct {
	TabID      string           `json:"tab_id"`
	FormData   memberParams     `json:"formData"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	Mode       string           `json:"mode"`
}

type modeParams struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type participantPaidAtParams struct {
	TabID                   string           `json:"tab_id"`
	TableQuery              utils.TableQuery `json:"tableQuery"`
	ParticipantPaidAtDialog struct {
		Value string `json:"value"`
	} `json:"participantPaidAtDialog"`
}

func memberIndexSignals(query map[string]any) map[string]any {
	return map[string]any{
		"tableQuery":     query,
		"mode":           "table",
		"formState":      "",
		"eventFormState": "",
		"editingId":      0,
		"formData":       map[string]any{"name": "", "description": ""},
		"errors":         map[string]any{"name": "", "description": ""},
	}
}

func memberShowSignals(data MemberData) map[string]any {
	return map[string]any{
		"mode":           "single",
		"tableQuery":     utils.TableQuerySignals(data.Query),
		"formState":      "",
		"eventFormState": "",
		"summaryMode":    data.Query.Summary,
		"formData": map[string]any{
			"name":        data.Member.Name,
			"description": data.Member.Description,
		},
		"errors": map[string]any{
			"name":        "",
			"description": "",
		},
		"participantPaidAtDialog": map[string]any{
			"open":        data.PaidAtDialog.Open,
			"fetching":    data.PaidAtDialog.Fetching,
			"title":       data.PaidAtDialog.Title,
			"message":     data.PaidAtDialog.Message,
			"eventId":     data.PaidAtDialog.EventID,
			"value":       data.PaidAtDialog.Value,
			"submitLabel": data.PaidAtDialog.SubmitLabel,
			"cancelLabel": data.PaidAtDialog.CancelLabel,
			"url":         data.PaidAtDialog.URL,
			"triggerID":   data.PaidAtDialog.TriggerID,
		},
	}
}
