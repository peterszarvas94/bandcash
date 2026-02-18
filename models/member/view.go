package member

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type MemberData struct {
	Title       string
	Member      *db.Member
	Events      []db.ListParticipantsByMemberRow
	Breadcrumbs []utils.Crumb
	GroupID     string
	IsAdmin     bool
}

type MembersData struct {
	Title       string
	Members     []db.Member
	Breadcrumbs []utils.Crumb
	GroupID     string
	IsAdmin     bool
}
