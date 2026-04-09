package store

type GetExpenseParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type CreateExpenseParams struct {
	ID          string      `json:"id"`
	GroupID     string      `json:"group_id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Date        string      `json:"date"`
	Paid        int64       `json:"paid"`
	PaidAt      interface{} `json:"paid_at"`
}

type UpdateExpenseParams struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Date        string      `json:"date"`
	Paid        int64       `json:"paid"`
	PaidAt      interface{} `json:"paid_at"`
	ID          string      `json:"id"`
	GroupID     string      `json:"group_id"`
}

type DeleteExpenseParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type ToggleExpensePaidParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}
