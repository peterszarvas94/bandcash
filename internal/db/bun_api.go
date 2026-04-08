package db

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

func GetUserByID(ctx context.Context, id string) (User, error) {
	var row User
	err := BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func GetGroupByID(ctx context.Context, id string) (Group, error) {
	var row Group
	err := BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func ListMembers(ctx context.Context, groupID string) ([]Member, error) {
	rows := make([]Member, 0)
	err := BunDB.NewSelect().Model(&rows).Where("group_id = ?", groupID).OrderExpr("created_at DESC").Scan(ctx)
	return rows, err
}

func GetEvent(ctx context.Context, arg GetEventParams) (Event, error) {
	var row Event
	err := BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func GetExpense(ctx context.Context, arg GetExpenseParams) (Expense, error) {
	var row Expense
	err := BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func GetMember(ctx context.Context, arg GetMemberParams) (Member, error) {
	var row Member
	err := BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func paidAtNullable(v interface{}) sql.NullString {
	switch t := v.(type) {
	case nil:
		return sql.NullString{}
	case string:
		t = strings.TrimSpace(t)
		if t == "" {
			return sql.NullString{}
		}
		return sql.NullString{String: t, Valid: true}
	case sql.NullString:
		if !t.Valid || strings.TrimSpace(t.String) == "" {
			return sql.NullString{}
		}
		return t
	default:
		return sql.NullString{}
	}
}

func currentTimestampNullString() sql.NullString {
	return sql.NullString{String: time.Now().UTC().Format("2006-01-02 15:04:05"), Valid: true}
}

func paidAtValue(v sql.NullString) interface{} {
	if v.Valid {
		return v.String
	}
	return nil
}

func eventTimeFromParts(date, eventTime string) string {
	date = strings.TrimSpace(date)
	eventTime = strings.TrimSpace(eventTime)
	if date == "" || eventTime == "" {
		return ""
	}
	return date + "T" + eventTime
}

func CreateEvent(ctx context.Context, arg CreateEventParams) (Event, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	if arg.Paid == 1 && !paidAt.Valid {
		paidAt = currentTimestampNullString()
	}

	event := Event{
		ID:          arg.ID,
		GroupID:     arg.GroupID,
		Title:       arg.Title,
		Time:        eventTimeFromParts(arg.Date, arg.EventTime),
		Date:        arg.Date,
		EventTime:   arg.EventTime,
		Place:       arg.Place,
		Description: arg.Description,
		Amount:      arg.Amount,
		Paid:        arg.Paid,
		PaidAt:      paidAt,
	}

	if _, err := BunDB.NewInsert().Model(&event).Exec(ctx); err != nil {
		return Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateEvent(ctx context.Context, arg UpdateEventParams) (Event, error) {
	current, err := GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return Event{}, err
	}

	paidAtInput := paidAtNullable(arg.PaidAt)
	finalPaidAt := current.PaidAt
	if arg.Paid == 0 {
		finalPaidAt = sql.NullString{}
	} else if paidAtInput.Valid {
		finalPaidAt = paidAtInput
	} else if current.Paid == 0 {
		finalPaidAt = currentTimestampNullString()
	}

	_, err = BunDB.NewUpdate().Model((*Event)(nil)).
		Set("title = ?", arg.Title).
		Set("time = ?", eventTimeFromParts(arg.Date, arg.EventTime)).
		Set("date = ?", arg.Date).
		Set("event_time = ?", arg.EventTime).
		Set("place = ?", arg.Place).
		Set("description = ?", arg.Description).
		Set("amount = ?", arg.Amount).
		Set("paid = ?", arg.Paid).
		Set("paid_at = ?", paidAtValue(finalPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteEvent(ctx context.Context, arg DeleteEventParams) error {
	_, err := BunDB.NewDelete().Model((*Event)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func ToggleEventPaid(ctx context.Context, arg ToggleEventPaidParams) (Event, error) {
	current, err := GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return Event{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = BunDB.NewUpdate().Model((*Event)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateEventPaidAt(ctx context.Context, arg UpdateEventPaidAtParams) (Event, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	paid := int64(0)
	if paidAt.Valid {
		paid = 1
	}

	_, err := BunDB.NewUpdate().Model((*Event)(nil)).
		Set("paid = ?", paid).
		Set("paid_at = ?", paidAtValue(paidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func CreateExpense(ctx context.Context, arg CreateExpenseParams) (Expense, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	if arg.Paid == 1 && !paidAt.Valid {
		paidAt = currentTimestampNullString()
	}

	expense := Expense{
		ID:          arg.ID,
		GroupID:     arg.GroupID,
		Title:       arg.Title,
		Description: arg.Description,
		Amount:      arg.Amount,
		Date:        arg.Date,
		Paid:        arg.Paid,
		PaidAt:      paidAt,
	}

	if _, err := BunDB.NewInsert().Model(&expense).Exec(ctx); err != nil {
		return Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateExpense(ctx context.Context, arg UpdateExpenseParams) (Expense, error) {
	current, err := GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return Expense{}, err
	}

	paidAtInput := paidAtNullable(arg.PaidAt)
	finalPaidAt := current.PaidAt
	if arg.Paid == 0 {
		finalPaidAt = sql.NullString{}
	} else if paidAtInput.Valid {
		finalPaidAt = paidAtInput
	} else if current.Paid == 0 {
		finalPaidAt = currentTimestampNullString()
	}

	_, err = BunDB.NewUpdate().Model((*Expense)(nil)).
		Set("title = ?", arg.Title).
		Set("description = ?", arg.Description).
		Set("amount = ?", arg.Amount).
		Set("date = ?", arg.Date).
		Set("paid = ?", arg.Paid).
		Set("paid_at = ?", paidAtValue(finalPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteExpense(ctx context.Context, arg DeleteExpenseParams) error {
	_, err := BunDB.NewDelete().Model((*Expense)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func ToggleExpensePaid(ctx context.Context, arg ToggleExpensePaidParams) (Expense, error) {
	current, err := GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return Expense{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = BunDB.NewUpdate().Model((*Expense)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
}

func CreateMember(ctx context.Context, arg CreateMemberParams) (Member, error) {
	member := Member{ID: arg.ID, GroupID: arg.GroupID, Name: arg.Name, Description: arg.Description}
	if _, err := BunDB.NewInsert().Model(&member).Exec(ctx); err != nil {
		return Member{}, err
	}
	return GetMember(ctx, GetMemberParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateMember(ctx context.Context, arg UpdateMemberParams) (Member, error) {
	_, err := BunDB.NewUpdate().Model((*Member)(nil)).
		Set("name = ?", arg.Name).
		Set("description = ?", arg.Description).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Member{}, err
	}
	return GetMember(ctx, GetMemberParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteMember(ctx context.Context, arg DeleteMemberParams) error {
	_, err := BunDB.NewDelete().Model((*Member)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func ToggleParticipantPaid(ctx context.Context, arg ToggleParticipantPaidParams) (Participant, error) {
	var current Participant
	err := BunDB.NewSelect().Model(&current).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	if err != nil {
		return Participant{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = BunDB.NewUpdate().Model((*Participant)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Participant{}, err
	}

	var p Participant
	err = BunDB.NewSelect().Model(&p).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	return p, err
}

func UpdateParticipantPaidAt(ctx context.Context, arg UpdateParticipantPaidAtParams) (Participant, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	paid := int64(0)
	if paidAt.Valid {
		paid = 1
	}

	_, err := BunDB.NewUpdate().Model((*Participant)(nil)).
		Set("paid = ?", paid).
		Set("paid_at = ?", paidAtValue(paidAt)).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return Participant{}, err
	}

	var p Participant
	err = BunDB.NewSelect().Model(&p).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	return p, err
}

func UpdateParticipantNote(ctx context.Context, arg UpdateParticipantNoteParams) error {
	_, err := BunDB.NewUpdate().Model((*Participant)(nil)).
		Set("note = ?", arg.Note).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	return err
}
