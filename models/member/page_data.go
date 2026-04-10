package member

import (
	"database/sql"

	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type MemberEvent struct {
	ID                 string
	GroupID            string
	Title              string
	Time               string
	Description        string
	Amount             int64
	ParticipantAmount  int64
	ParticipantExpense int64
	ParticipantPaid    int64
	ParticipantPaidAt  sql.NullString
}

type MemberData struct {
	Title           string
	Member          *db.Member
	Events          []MemberEvent
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	GroupID         string
	IsAdmin         bool
	IsAuthenticated bool
	IsSuperAdmin    bool
	Query           utils.TableQuery
	Pager           utils.TablePagination
	RecentYears     []int
	TotalCut        int64
	TotalExpense    int64
	TotalPayout     int64
	TotalPaid       int64
	TotalUnpaid     int64
	EventsTable     utils.TableLayout
	PaidAtDialog    ParticipantPaidAtDialogState
}

type ParticipantPaidAtDialogState struct {
	Open        bool
	Fetching    bool
	Title       string
	Message     string
	EventID     string
	Value       string
	SubmitLabel string
	CancelLabel string
	URL         string
	TriggerID   string
}

type MembersData struct {
	Title           string
	GroupName       string
	Members         []MemberListRow
	Query           utils.TableQuery
	Pager           utils.TablePagination
	Breadcrumbs     []utils.Crumb
	Signals         map[string]any
	GroupID         string
	IsAdmin         bool
	IsAuthenticated bool
	IsSuperAdmin    bool
	MembersTable    utils.TableLayout
}

type MemberListRow struct {
	ID          string
	Name        string
	Description string
	Unpaid      int64
}

type NewMemberPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}

type EditMemberPageData struct {
	Title           string
	Breadcrumbs     []utils.Crumb
	GroupID         string
	Member          *db.Member
	Signals         map[string]any
	IsAuthenticated bool
	IsSuperAdmin    bool
}
