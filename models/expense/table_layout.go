package expense

import "bandcash/internal/utils"

func ExpensesIndexTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "date"},
		{Key: "amount"},
		{Key: "paid", MaxWRem: 7, WRem: 7},
		{Key: "paid_at", MaxWRem: 12, WRem: 12},
	}, 0)
}
