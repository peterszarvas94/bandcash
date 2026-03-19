package member

import (
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
}

type MemberData struct {
	Title        string
	Member       *db.Member
	Events       []MemberEvent
	Breadcrumbs  []utils.Crumb
	GroupID      string
	IsAdmin      bool
	UserEmail    string
	Query        utils.TableQuery
	Pager        utils.TablePagination
	RecentYears  []int
	TotalCut     int64
	TotalExpense int64
	TotalPayout  int64
	EventsTable  utils.TableLayout
}

type MembersData struct {
	Title        string
	Members      []db.Member
	Query        utils.TableQuery
	Pager        utils.TablePagination
	Breadcrumbs  []utils.Crumb
	GroupID      string
	IsAdmin      bool
	UserEmail    string
	MembersTable utils.TableLayout
}
