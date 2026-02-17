package payee

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type PayeeData struct {
	Title       string
	Payee       *db.Payee
	Entries     []db.ListParticipantsByPayeeRow
	Breadcrumbs []utils.Crumb
}

type PayeesData struct {
	Title       string
	Payees      []db.Payee
	Breadcrumbs []utils.Crumb
}
