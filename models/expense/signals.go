package expense

import "bandcash/internal/utils"

type expenseParams struct {
	Title       string `json:"title" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Date        string `json:"date" validate:"required"`
	Paid        bool   `json:"paid"`
	PaidAt      string `json:"paidAt"`
}

type expenseTableParams struct {
	TabID      string           `json:"tab_id"`
	FormData   expenseParams    `json:"formData"`
	TableQuery utils.TableQuery `json:"tableQuery"`
	Mode       string           `json:"mode"`
}

type modeParams struct {
	TabID      string           `json:"tab_id"`
	Mode       string           `json:"mode"`
	TableQuery utils.TableQuery `json:"tableQuery"`
}

type paidAtParams struct {
	TabID        string           `json:"tab_id"`
	Mode         string           `json:"mode"`
	TableQuery   utils.TableQuery `json:"tableQuery"`
	PaidAtDialog struct {
		Value string `json:"value"`
	} `json:"paidAtDialog"`
}

func expenseIndexSignals(query utils.TableQuery) map[string]any {
	return map[string]any{
		"tableQuery":      utils.TableQuerySignals(query),
		"dateRange":       map[string]any{"from": query.From, "to": query.To},
		"showCustomRange": query.DateMode == "custom" || query.From != "" || query.To != "",
		"mode":            "table",
		"formState":       "",
		"eventFormState":  "",
		"summaryMode":     query.Summary,
		"editingId":       0,
		"paidAtDialog": map[string]any{
			"open":        false,
			"fetching":    false,
			"mode":        "table",
			"title":       "",
			"message":     "",
			"value":       "",
			"placeholder": "",
			"submitLabel": "",
			"cancelLabel": "",
			"url":         "",
			"triggerID":   "",
		},
		"formData": map[string]any{"title": "", "description": "", "amount": 0, "date": "", "paid": false, "paidAt": ""},
		"errors":   map[string]any{"title": "", "description": "", "amount": "", "date": ""},
	}
}

func expenseShowSignals(data ExpenseData) map[string]any {
	return map[string]any{
		"mode": "single",
		"paidAtDialog": map[string]any{
			"open":        data.PaidAtDialog.Open,
			"fetching":    data.PaidAtDialog.Fetching,
			"mode":        data.PaidAtDialog.Mode,
			"title":       data.PaidAtDialog.Title,
			"message":     data.PaidAtDialog.Message,
			"value":       data.PaidAtDialog.Value,
			"placeholder": data.PaidAtDialog.Placeholder,
			"submitLabel": data.PaidAtDialog.SubmitLabel,
			"cancelLabel": data.PaidAtDialog.CancelLabel,
			"url":         data.PaidAtDialog.URL,
			"triggerID":   data.PaidAtDialog.TriggerID,
		},
	}
}
