package event

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EventData struct {
	Title                   string
	Event                   *db.Event
	Participants            []db.ListParticipantsByEventRow
	WizardRows              []ParticipantWizardRow
	Query                   utils.TableQuery
	Pager                   utils.TablePagination
	Members                 []db.Member
	AllMembers              []db.Member
	Breadcrumbs             []utils.Crumb
	Signals                 map[string]any
	Leftover                int64
	TotalPaid               int64
	TotalUnpaid             int64
	FilteredPaid            int64
	FilteredUnpaid          int64
	FilteredLeftover        int64
	WizardEventAmount       int64
	WizardError             string
	EditorMode              string
	GroupID                 string
	IsAdmin                 bool
	PaidAtDialog            PaidAtDialogState
	ParticipantPaidAtDialog ParticipantPaidAtDialogState
	ParticipantNoteDialog   ParticipantNoteDialogState
	IsAuthenticated         bool
	IsSuperAdmin            bool
	ParticipantsTable       utils.TableLayout
}

type PaidAtDialogState struct {
	Open        bool
	Fetching    bool
	Mode        string
	Title       string
	Message     string
	Value       string
	Placeholder string
	SubmitLabel string
	CancelLabel string
	URL         string
	TriggerID   string
}

type ParticipantNoteDialogState struct {
	Open        bool
	Fetching    bool
	ReadOnly    bool
	Title       string
	Message     string
	MemberID    string
	Value       string
	SubmitLabel string
	CancelLabel string
	URL         string
	TriggerID   string
}

type ParticipantPaidAtDialogState struct {
	Open        bool
	Fetching    bool
	Title       string
	Message     string
	MemberID    string
	Value       string
	SubmitLabel string
	CancelLabel string
	URL         string
	TriggerID   string
}

type ParticipantWizardRow struct {
	RowID      string
	MemberID   string
	MemberName string
	Included   bool
	Amount     int64
	Expense    int64
	Note       string
	Paid       bool
	PaidAt     string
}

type NewEventPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type EventsData struct {
	Title                  string
	GroupName              string
	GroupAdminEmail        string
	GroupCreatedAt         string
	Events                 []db.Event
	RecentYears            []int
	Query                  utils.TableQuery
	Pager                  utils.TablePagination
	Breadcrumbs            []utils.Crumb
	Signals                map[string]any
	GroupID                string
	IsAdmin                bool
	IsAuthenticated        bool
	IsSuperAdmin           bool
	TotalEventAmount       int64
	TotalPaid              int64
	TotalUnpaid            int64
	FilteredTotal          int64
	FilteredIncomePaid     int64
	FilteredIncomeUnpaid   int64
	FilteredPayoutsPaid    int64
	FilteredPayoutsUnpaid  int64
	FilteredExpensesPaid   int64
	FilteredExpensesUnpaid int64
	EventsTable            utils.TableLayout
	PaidAtDialog           PaidAtDialogState
}
