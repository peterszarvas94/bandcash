package utils

func EventsIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title", WidthCh: 28},
		{Key: "time", WidthCh: 19},
		{Key: "amount", WidthCh: 12},
		{Key: "paid", WidthCh: 14},
		{Key: "paid_at", WidthCh: 14},
	}, 4)
}

func EventParticipantsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name", WidthCh: 24},
		{Key: "amount", WidthCh: 12},
		{Key: "expense", WidthCh: 12},
		{Key: "total", WidthCh: 12},
		{Key: "note", WidthCh: 24},
		{Key: "paid", WidthCh: 14},
		{Key: "paid_at", WidthCh: 20},
	}, 0)
}

func ExpensesIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title", WidthCh: 20},
		{Key: "date", WidthCh: 12},
		{Key: "amount", WidthCh: 12},
		{Key: "paid", WidthCh: 14},
		{Key: "paid_at", WidthCh: 14},
	}, 4)
}

func GroupsIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name", WidthCh: 24},
		{Key: "role", WidthCh: 10},
	}, 4)
}

func MembersIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name", WidthCh: 20},
		{Key: "description", WidthCh: 32},
	}, 4)
}

func MemberEventsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title", WidthCh: 22},
		{Key: "time", WidthCh: 19},
		{Key: "participant_amount", WidthCh: 12},
		{Key: "participant_expense", WidthCh: 12},
		{Key: "total", WidthCh: 12},
		{Key: "paid", WidthCh: 14},
		{Key: "paid_at", WidthCh: 14},
	}, 0)
}

func ViewersAdminsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email", WidthCh: 34},
	}, 2)
}

func ViewersPendingTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email", WidthCh: 34},
		{Key: "role", WidthCh: 10},
		{Key: "created_at", WidthCh: 18},
	}, 2)
}

func ViewersTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email", WidthCh: 34},
	}, 2)
}

func GroupAccessTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email", WidthCh: 30},
		{Key: "role", WidthCh: 10},
		{Key: "status", WidthCh: 10},
	}, 2)
}

func AdminUsersTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email", WidthCh: 34},
		{Key: "created_at", WidthCh: 18},
		{Key: "status", WidthCh: 12},
	}, 2)
}

func AdminGroupsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name", WidthCh: 24},
		{Key: "admin", WidthCh: 24},
		{Key: "viewers", WidthCh: 10},
		{Key: "created_at", WidthCh: 18},
	}, 0)
}
