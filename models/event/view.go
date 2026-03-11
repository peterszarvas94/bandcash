package event

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EventData struct {
	Title             string
	Event             *db.Event
	Participants      []db.ListParticipantsByEventRow
	Query             utils.TableQuery
	Pager             utils.TablePagination
	Members           []db.Member
	Breadcrumbs       []utils.Crumb
	Leftover          int64
	TotalDistributed  int64
	GroupID           string
	IsAdmin           bool
	UserEmail         string
	ParticipantsTable utils.TableLayout
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
	EventsTable      utils.TableLayout
}
