package event

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EventData struct {
	Title            string
	Event            *db.Event
	Participants     []db.ListParticipantsByEventRow
	Members          []db.Member
	MemberIDs        map[string]bool
	Breadcrumbs      []utils.Crumb
	Leftover         int64
	TotalDistributed int64
	GroupID          string
	IsAdmin          bool
	UserEmail        string
}

type EventsData struct {
	Title       string
	Events      []db.Event
	Breadcrumbs []utils.Crumb
	GroupID     string
	GroupName   string
	IsAdmin     bool
	UserEmail   string
}
