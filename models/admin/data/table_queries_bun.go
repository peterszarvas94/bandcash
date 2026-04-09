package data

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"bandcash/internal/db"
)

type AdminUserTableRow struct {
	ID        string       `bun:"id"`
	Email     string       `bun:"email"`
	CreatedAt sql.NullTime `bun:"created_at"`
	IsBanned  int64        `bun:"is_banned"`
}

type AdminSessionTableRow struct {
	ID        string       `bun:"id"`
	UserID    string       `bun:"user_id"`
	UserEmail string       `bun:"user_email"`
	CreatedAt sql.NullTime `bun:"created_at"`
	ExpiresAt time.Time    `bun:"expires_at"`
}

func CountUsersTable(ctx context.Context, search string) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("users")
	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("users.email LIKE '%' || ? || '%'", search)
	}
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListUsersTable(ctx context.Context, search, sort, dir string, limit, offset int) ([]AdminUserTableRow, error) {
	rows := make([]AdminUserTableRow, 0)
	q := db.BunDB.NewSelect().
		TableExpr("users").
		ColumnExpr("users.id").
		ColumnExpr("users.email").
		ColumnExpr("users.created_at").
		ColumnExpr("CAST(CASE WHEN bu.user_id IS NULL THEN 0 ELSE 1 END AS INTEGER) AS is_banned").
		Join("LEFT JOIN banned_users AS bu ON bu.user_id = users.id")

	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("users.email LIKE '%' || ? || '%'", search)
	}

	d := normalizeDir(dir)
	switch sort {
	case "email":
		q = q.OrderExpr("users.email " + d)
	case "createdAt":
		q = q.OrderExpr("users.created_at " + d)
	default:
		q = q.OrderExpr("users.created_at DESC")
	}
	q = q.OrderExpr("users.id ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	err := q.Scan(ctx, &rows)
	return rows, err
}

func CountGroupsTable(ctx context.Context, search string) (int64, error) {
	q := db.BunDB.NewSelect().TableExpr("groups")
	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("name LIKE '%' || ? || '%'", search)
	}
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListGroupsTable(ctx context.Context, search, sort, dir string, limit, offset int) ([]db.Group, error) {
	rows := make([]db.Group, 0)
	q := db.BunDB.NewSelect().Model(&rows)

	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("name LIKE '%' || ? || '%'", search)
	}

	d := normalizeDir(dir)
	switch sort {
	case "name":
		q = q.OrderExpr("name " + d)
	case "createdAt":
		q = q.OrderExpr("created_at " + d)
	default:
		q = q.OrderExpr("created_at DESC")
	}
	q = q.OrderExpr("id ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	err := q.Scan(ctx)
	return rows, err
}

func CountSessionsTable(ctx context.Context, search string) (int64, error) {
	q := db.BunDB.NewSelect().
		TableExpr("user_sessions").
		Join("JOIN users ON users.id = user_sessions.user_id")
	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("users.email LIKE '%' || ? || '%'", search)
	}
	n, err := q.Count(ctx)
	return int64(n), err
}

func ListSessionsTable(ctx context.Context, search, sort, dir string, limit, offset int) ([]AdminSessionTableRow, error) {
	rows := make([]AdminSessionTableRow, 0)
	q := db.BunDB.NewSelect().
		TableExpr("user_sessions").
		ColumnExpr("user_sessions.id").
		ColumnExpr("user_sessions.user_id").
		ColumnExpr("users.email AS user_email").
		ColumnExpr("user_sessions.created_at").
		ColumnExpr("user_sessions.expires_at").
		Join("JOIN users ON users.id = user_sessions.user_id")

	search = strings.TrimSpace(search)
	if search != "" {
		q = q.Where("users.email LIKE '%' || ? || '%'", search)
	}

	d := normalizeDir(dir)
	switch sort {
	case "email":
		q = q.OrderExpr("users.email " + d)
	case "createdAt":
		q = q.OrderExpr("user_sessions.created_at " + d)
	default:
		q = q.OrderExpr("user_sessions.created_at DESC")
	}
	q = q.OrderExpr("user_sessions.id ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	err := q.Scan(ctx, &rows)
	return rows, err
}

func normalizeDir(dir string) string {
	if strings.EqualFold(strings.TrimSpace(dir), "asc") {
		return "ASC"
	}
	return "DESC"
}
