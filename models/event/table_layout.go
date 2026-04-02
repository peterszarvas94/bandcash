package event

import "bandcash/internal/utils"

func EventsIndexTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "time"},
		{Key: "place"},
		{Key: "amount"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}

func EventParticipantsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "name", MaxWRem: 12},
		{Key: "amount", MaxWRem: 8},
		{Key: "expense", MaxWRem: 8},
		{Key: "total"},
		{Key: "paid"},
		{Key: "paid_at", MaxWRem: 8},
		{Key: "note", MaxWRem: 24},
	}, 0)
}
