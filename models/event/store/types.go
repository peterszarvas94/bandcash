package store

import "database/sql"

type GetEventParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type CreateEventParams struct {
	ID          string      `json:"id"`
	GroupID     string      `json:"group_id"`
	Title       string      `json:"title"`
	Date        string      `json:"date"`
	EventTime   string      `json:"event_time"`
	Place       string      `json:"place"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Paid        int64       `json:"paid"`
	PaidAt      interface{} `json:"paid_at"`
}

type UpdateEventParams struct {
	Title       string      `json:"title"`
	Date        string      `json:"date"`
	EventTime   string      `json:"event_time"`
	Place       string      `json:"place"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Paid        int64       `json:"paid"`
	PaidAt      interface{} `json:"paid_at"`
	ID          string      `json:"id"`
	GroupID     string      `json:"group_id"`
}

type DeleteEventParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type ToggleEventPaidParams struct {
	ID      string `json:"id"`
	GroupID string `json:"group_id"`
}

type UpdateEventPaidAtParams struct {
	PaidAt  interface{} `json:"paid_at"`
	ID      string      `json:"id"`
	GroupID string      `json:"group_id"`
}

type ToggleParticipantPaidParams struct {
	EventID  string `json:"event_id"`
	MemberID string `json:"member_id"`
	GroupID  string `json:"group_id"`
}

type UpdateParticipantPaidAtParams struct {
	PaidAt   interface{} `json:"paid_at"`
	EventID  string      `json:"event_id"`
	MemberID string      `json:"member_id"`
	GroupID  string      `json:"group_id"`
}

type UpdateParticipantNoteParams struct {
	Note     string `json:"note"`
	EventID  string `json:"event_id"`
	MemberID string `json:"member_id"`
	GroupID  string `json:"group_id"`
}

type UpdateParticipantParams struct {
	Amount   int64       `json:"amount"`
	Expense  int64       `json:"expense"`
	Note     string      `json:"note"`
	Paid     int64       `json:"paid"`
	PaidAt   interface{} `json:"paid_at"`
	EventID  string      `json:"event_id"`
	MemberID string      `json:"member_id"`
	GroupID  string      `json:"group_id"`
}

type AddParticipantParams struct {
	GroupID  string      `json:"group_id"`
	EventID  string      `json:"event_id"`
	MemberID string      `json:"member_id"`
	Amount   int64       `json:"amount"`
	Expense  int64       `json:"expense"`
	Note     string      `json:"note"`
	Paid     int64       `json:"paid"`
	PaidAt   interface{} `json:"paid_at"`
}

type RemoveParticipantParams struct {
	EventID  string `json:"event_id"`
	MemberID string `json:"member_id"`
	GroupID  string `json:"group_id"`
}

type ListParticipantsByEventParams struct {
	EventID string `json:"event_id"`
	GroupID string `json:"group_id"`
}

type ListParticipantsByEventRow struct {
	ID                 string         `json:"id"`
	GroupID            string         `json:"group_id"`
	Name               string         `json:"name"`
	Description        string         `json:"description"`
	CreatedAt          sql.NullTime   `json:"created_at"`
	UpdatedAt          sql.NullTime   `json:"updated_at"`
	ParticipantAmount  int64          `json:"participant_amount"`
	ParticipantExpense int64          `json:"participant_expense"`
	ParticipantNote    string         `json:"participant_note"`
	ParticipantPaid    int64          `json:"participant_paid"`
	ParticipantPaidAt  sql.NullString `json:"participant_paid_at"`
}

type SumParticipantPaidAmountsByGroupRow struct {
	PaidAmount   int64 `json:"paid_amount"`
	UnpaidAmount int64 `json:"unpaid_amount"`
}
