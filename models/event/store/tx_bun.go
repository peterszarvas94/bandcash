package store

import (
	"context"
	"database/sql"

	"bandcash/internal/db"
	"github.com/uptrace/bun"
)

func getEventTx(ctx context.Context, tx bun.Tx, arg GetEventParams) (db.Event, error) {
	var row db.Event
	err := tx.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func UpdateEventTx(ctx context.Context, tx bun.Tx, arg UpdateEventParams) (db.Event, error) {
	current, err := getEventTx(ctx, tx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
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

	_, err = tx.NewUpdate().Model((*db.Event)(nil)).
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
	return getEventTx(ctx, tx, GetEventParams{ID: arg.ID, GroupID: arg.GroupID})
}

func ListParticipantsByEventTx(ctx context.Context, tx bun.Tx, arg ListParticipantsByEventParams) ([]ListParticipantsByEventRow, error) {
	rows := make([]ListParticipantsByEventRow, 0)
	err := tx.NewSelect().
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

func UpdateParticipantTx(ctx context.Context, tx bun.Tx, arg UpdateParticipantParams) error {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE participants
SET amount = ?,
    expense = ?,
    note = ?,
    paid = ?,
    paid_at = CASE
      WHEN ? = 0 THEN NULL
      WHEN ? IS NOT NULL THEN ?
      WHEN paid = 0 THEN CURRENT_TIMESTAMP
      ELSE paid_at
    END
WHERE event_id = ?
  AND member_id = ?
  AND group_id = ?`,
		arg.Amount,
		arg.Expense,
		arg.Note,
		arg.Paid,
		arg.Paid,
		arg.PaidAt,
		arg.PaidAt,
		arg.EventID,
		arg.MemberID,
		arg.GroupID,
	)
	return err
}

func AddParticipantTx(ctx context.Context, tx bun.Tx, arg AddParticipantParams) (db.Participant, error) {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO participants (group_id, event_id, member_id, amount, expense, note, paid, paid_at)
VALUES (?, ?, ?, ?, ?, ?, ?, CASE WHEN ? = 0 THEN NULL ELSE ? END)`,
		arg.GroupID,
		arg.EventID,
		arg.MemberID,
		arg.Amount,
		arg.Expense,
		arg.Note,
		arg.Paid,
		arg.Paid,
		arg.PaidAt,
	)
	if err != nil {
		return db.Participant{}, err
	}

	var p db.Participant
	err = tx.NewSelect().Model(&p).
		Where("group_id = ?", arg.GroupID).
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Scan(ctx)
	return p, err
}

func RemoveParticipantTx(ctx context.Context, tx bun.Tx, arg RemoveParticipantParams) error {
	_, err := tx.NewDelete().
		TableExpr("participants").
		Where("event_id = ?", arg.EventID).
		Where("member_id = ?", arg.MemberID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	return err
}
