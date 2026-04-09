package store

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"bandcash/internal/db"
)

func GetExpense(ctx context.Context, arg GetExpenseParams) (db.Expense, error) {
	var row db.Expense
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Scan(ctx)
	return row, err
}

func GetExpenseByID(ctx context.Context, id string) (db.Expense, error) {
	var row db.Expense
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func ListExpenses(ctx context.Context, groupID string) ([]db.Expense, error) {
	rows := make([]db.Expense, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("group_id = ?", groupID).
		OrderExpr("date DESC").
		OrderExpr("created_at DESC").
		Scan(ctx)
	return rows, err
}

func CreateExpense(ctx context.Context, arg CreateExpenseParams) (db.Expense, error) {
	paidAt := paidAtNullable(arg.PaidAt)
	if arg.Paid == 1 && !paidAt.Valid {
		paidAt = currentTimestampNullString()
	}

	expense := db.Expense{
		ID:          arg.ID,
		GroupID:     arg.GroupID,
		Title:       arg.Title,
		Description: arg.Description,
		Amount:      arg.Amount,
		Date:        arg.Date,
		Paid:        arg.Paid,
		PaidAt:      paidAt,
	}

	if _, err := db.BunDB.NewInsert().Model(&expense).Exec(ctx); err != nil {
		return db.Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
}

func UpdateExpense(ctx context.Context, arg UpdateExpenseParams) (db.Expense, error) {
	current, err := GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return db.Expense{}, err
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

	_, err = db.BunDB.NewUpdate().Model((*db.Expense)(nil)).
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
		return db.Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
}

func DeleteExpense(ctx context.Context, arg DeleteExpenseParams) error {
	_, err := db.BunDB.NewDelete().Model((*db.Expense)(nil)).Where("id = ?", arg.ID).Where("group_id = ?", arg.GroupID).Exec(ctx)
	return err
}

func DeleteExpenseByID(ctx context.Context, id string) error {
	_, err := db.BunDB.NewDelete().Model((*db.Expense)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func ToggleExpensePaid(ctx context.Context, arg ToggleExpensePaidParams) (db.Expense, error) {
	current, err := GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
	if err != nil {
		return db.Expense{}, err
	}

	nextPaid := int64(1)
	nextPaidAt := currentTimestampNullString()
	if current.Paid == 1 {
		nextPaid = 0
		nextPaidAt = sql.NullString{}
	}

	_, err = db.BunDB.NewUpdate().Model((*db.Expense)(nil)).
		Set("paid = ?", nextPaid).
		Set("paid_at = ?", paidAtValue(nextPaidAt)).
		Where("id = ?", arg.ID).
		Where("group_id = ?", arg.GroupID).
		Exec(ctx)
	if err != nil {
		return db.Expense{}, err
	}
	return GetExpense(ctx, GetExpenseParams{ID: arg.ID, GroupID: arg.GroupID})
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
