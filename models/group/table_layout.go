package group

import "bandcash/internal/utils"

func GroupUsersTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "email"},
		{Key: "role"},
		{Key: "status"},
	}, 0)
}

func GroupPaymentsEventsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}

func GroupPaymentsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "name"},
		{Key: "type"},
		{Key: "date"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}

func GroupPaymentsParticipantsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "event"},
		{Key: "member"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}

func GroupPaymentsExpensesTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}
