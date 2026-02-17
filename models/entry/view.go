package entry

import (
	"bandcash/internal/db"
	"bandcash/internal/utils"
)

type EntryData struct {
	Title            string
	Entry            *db.Entry
	Participants     []db.ListParticipantsByEntryRow
	Payees           []db.Payee
	PayeeIDs         map[int64]bool
	Breadcrumbs      []utils.Crumb
	Leftover         int64
	TotalDistributed int64
}

type EntriesData struct {
	Title       string
	Entries     []db.Entry
	Breadcrumbs []utils.Crumb
}
