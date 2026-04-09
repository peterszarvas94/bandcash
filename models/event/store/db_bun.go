package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"bandcash/internal/db"
)

func GetEvent(ctx context.Context, arg GetEventParams) (db.Event, error) {
	var row db.Event
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func GetEventByID(ctx context.Context, id string) (db.Event, error) {
	var row db.Event
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func ListEvents(ctx context.Context, groupID string) ([]db.Event, error) {
	rows := make([]db.Event, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("group_id = ?", groupID).
		OrderExpr("time ASC").
		Scan(ctx)
	return rows, err
}

func ListPaidEventsByGroup(ctx context.Context, groupID string) ([]db.Event, error) {
	rows := make([]db.Event, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("group_id = ?", groupID).
		Where("paid = 1").
		OrderExpr("COALESCE(paid_at, updated_at) DESC").
		OrderExpr("updated_at DESC").
		Scan(ctx)
	return rows, err
}

func ListUnpaidEventsByGroup(ctx context.Context, groupID string) ([]db.Event, error) {
	rows := make([]db.Event, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("group_id = ?", groupID).
		Where("paid = 0").
		OrderExpr("time ASC").
		OrderExpr("created_at DESC").
		Scan(ctx)
	return rows, err
}

func CreateEvent(ctx context.Context, arg CreateEventParams) (db.Event, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	if arg.Paid == 1 && !paidAt.Valid {
		paidAt = currentTimestampNullString()
	}

	event := db.Event{
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

	if _, err := db.BunDB.NewInsert().Model(&event).Exec(ctx); err != nil {
		return db.Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateEvent(ctx context.Context, arg UpdateEventParams) (db.Event, error) {
	current, err := GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return db.Event{}, err
	}

	paidAtInput := paidAtNullable(arg.PaidAt)
	finalPaidAt := current.PaidAt
	if arg.Paid == 0 || isPaidAtClearRequest(arg.PaidAt) {
		finalPaidAt = sql.NullString{}
	} else if paidAtInput.Valid {
		finalPaidAt = paidAtInput
	} else if current.Paid == 0 {
		finalPaidAt = currentTimestampNullString()
	}

	_, err = db.BunDB.NewUpdate().Model((*db.Event)(nil)).
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
		return db.Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteEvent(ctx context.Context, arg DeleteEventParams) error {
	_, err := db.BunDB.NewDelete().Model((*db.Event)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func DeleteEventByID(ctx context.Context, id string) error {
	_, err := db.BunDB.NewDelete().Model((*db.Event)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func ToggleEventPaid(ctx context.Context, arg ToggleEventPaidParams) (db.Event, error) {
	current, err := GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return db.Event{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = db.BunDB.NewUpdate().Model((*db.Event)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateEventPaidAt(ctx context.Context, arg UpdateEventPaidAtParams) (db.Event, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	paid := int64(0)
	if paidAt.Valid {
		paid = 1
	}

	_, err := db.BunDB.NewUpdate().Model((*db.Event)(nil)).
		Set("paid = ?", paid).
		Set("paid_at = ?", paidAtValue(paidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Event{}, err
	}
	return GetEvent(ctx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func ListParticipantsByEvent(ctx context.Context, arg ListParticipantsByEventParams) ([]ListParticipantsByEventRow, error) {
	rows := make([]ListParticipantsByEventRow, 0)
	err := db.BunDB.NewSelect().
		TableExpr("members").
		ColumnExpr("members.id").
		ColumnExpr("members.group_id").
		ColumnExpr("members.name").
		ColumnExpr("members.description").
		ColumnExpr("members.created_at").
		ColumnExpr("members.updated_at").
		ColumnExpr("participants.amount AS participant_amount").
		ColumnExpr("participants.expense AS participant_expense").
		ColumnExpr("participants.note AS participant_note").
		ColumnExpr("participants.paid AS participant_paid").
		ColumnExpr("participants.paid_at AS participant_paid_at").
		Join("JOIN participants ON participants.member_id = members.id").
		Where("participants.event_id = ?", arg.EventID).
		Where("participants.group_id = ?", arg.GroupID).
		OrderExpr("members.name ASC").
		Scan(ctx, &rows)
	return rows, err
}

func SumParticipantPaidAmountsByGroup(ctx context.Context, groupID string) (SumParticipantPaidAmountsByGroupRow, error) {
	var row SumParticipantPaidAmountsByGroupRow
	err := db.BunDB.NewSelect().
		TableExpr("participants").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid_amount").
		ColumnExpr("CAST(COALESCE(SUM(CASE WHEN paid = 0 THEN amount ELSE 0 END), 0) AS INTEGER) AS unpaid_amount").
		Where("group_id = ?", groupID).
		Scan(ctx, &row)
	return row, err
}

func ToggleParticipantPaid(ctx context.Context, arg ToggleParticipantPaidParams) (db.Participant, error) {
	var current db.Participant
	err := db.BunDB.NewSelect().Model(&current).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	if err != nil {
		return db.Participant{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = db.BunDB.NewUpdate().Model((*db.Participant)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Participant{}, err
	}

	var p db.Participant
	err = db.BunDB.NewSelect().Model(&p).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	return p, err
}

func UpdateParticipantPaidAt(ctx context.Context, arg UpdateParticipantPaidAtParams) (db.Participant, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	paid := int64(0)
	if paidAt.Valid {
		paid = 1
	}

	_, err := db.BunDB.NewUpdate().Model((*db.Participant)(nil)).
		Set("paid = ?", paid).
		Set("paid_at = ?", paidAtValue(paidAt)).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Participant{}, err
	}

	var p db.Participant
	err = db.BunDB.NewSelect().Model(&p).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Scan(ctx)
	return p, err
}

func UpdateParticipantNote(ctx context.Context, arg UpdateParticipantNoteParams) error {
	_, err := db.BunDB.NewUpdate().Model((*db.Participant)(nil)).
		Set("note = ?", arg.Note).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	return err
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

func isPaidAtClearRequest(v interface{}) bool {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case sql.NullString:
		return t.Valid && strings.TrimSpace(t.String) == ""
	default:
		return false
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
