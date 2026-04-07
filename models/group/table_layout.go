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
	}, 0)
}

func GroupPaymentsParticipantsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "member"},
		{Key: "amount"},
	}, 0)
}

func GroupPaymentsExpensesTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "amount"},
	}, 0)
}
