package event

import "bandcash/internal/utils"

func EventsIndexTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "time"},
		{Key: "place"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}

func EventParticipantsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "name", MaxWRem: 12},
		{Key: "amount", MaxWRem: 8},
		{Key: "expense", MaxWRem: 8},
		{Key: "total"},
		{Key: "note", MaxWRem: 16},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}
