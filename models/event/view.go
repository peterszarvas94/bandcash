package event

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EventData struct {
	Title                string
	Event                *db.Event
	Participants         []db.ListParticipantsByEventRow
	WizardRows           []ParticipantWizardRow
	Query                utils.TableQuery
	Pager                utils.TablePagination
	Members              []db.Member
	AllMembers           []db.Member
	WizardAddableMembers []db.Member
	Breadcrumbs          []utils.Crumb
	Leftover             int64
	TotalPaid            int64
	TotalUnpaid          int64
	WizardEventAmount    int64
	WizardError          string
	EditorMode           string
	GroupID              string
	IsAdmin              bool
	UserEmail            string
	ParticipantsTable    utils.TableLayout
}

type ParticipantWizardRow struct {
	MemberID   string
	MemberName string
	Included   bool
	Amount     int64
	Expense    int64
}

type EventsData struct {
	Title            string
	Events           []db.Event
	RecentYears      []int
	Query            utils.TableQuery
	Pager            utils.TablePagination
	Breadcrumbs      []utils.Crumb
	GroupID          string
	IsAdmin          bool
	UserEmail        string
	TotalEventAmount int64
	FilteredTotal    int64
	FilteredPaid     int64
	FilteredUnpaid   int64
	EventsTable      utils.TableLayout
}
