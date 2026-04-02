package utils

func EventsIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title"},
		{Key: "time"},
		{Key: "place"},
		{Key: "amount"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}

func EventParticipantsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name", MaxWRem: 10},
		{Key: "amount", MaxWRem: 8},
		{Key: "expense", MaxWRem: 8},
		{Key: "total"},
		{Key: "paid"},
		{Key: "paid_at", MaxWRem: 8},
		{Key: "note", MaxWRem: 24},
	}, 0)
}

func ExpensesIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title"},
		{Key: "date"},
		{Key: "amount"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}

func GroupsIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name"},
		{Key: "role"},
	}, 0)
}

func MembersIndexTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name"},
		{Key: "description"},
	}, 0)
}

func MemberEventsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "title"},
		{Key: "time"},
		{Key: "participant_amount"},
		{Key: "participant_expense"},
		{Key: "total"},
		{Key: "paid"},
		{Key: "paid_at"},
	}, 0)
}

func ViewersAdminsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email"},
	}, 2)
}

func ViewersPendingTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email"},
		{Key: "role"},
		{Key: "created_at"},
	}, 2)
}

func ViewersTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email"},
	}, 2)
}

func GroupUsersTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email"},
		{Key: "role"},
		{Key: "status"},
	}, 0)
}

func AdminUsersTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "email"},
		{Key: "created_at"},
		{Key: "status"},
	}, 2)
}

func AdminGroupsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "name"},
		{Key: "admin"},
		{Key: "viewers"},
		{Key: "created_at"},
	}, 0)
}

func AdminSessionsTableLayout() TableLayout {
	return NewTableLayout([]TableColumn{
		{Key: "user_email"},
		{Key: "session_id"},
		{Key: "created_at"},
		{Key: "expires_at"},
		{Key: "actions"},
	}, 0)
}
