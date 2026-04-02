package member

import "bandcash/internal/utils"

func MembersIndexTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "name"},
		{Key: "description"},
	}, 0)
}

func MemberEventsTableLayout() utils.TableLayout {
	return utils.NewTableLayout([]utils.TableColumn{
		{Key: "title"},
		{Key: "time"},
		{Key: "participant_amount"},
		{Key: "participant_expense"},
		{Key: "total"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}
