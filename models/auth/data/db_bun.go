package data

import (
	"context"
	"strings"

	"bandcash/internal/db"
)

func GetUserByID(ctx context.Context, id string) (db.User, error) {
	var row db.User
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	return row, err
}

func BanUser(ctx context.Context, arg BanUserParams) error {
	_, err := db.BunDB.ExecContext(ctx, "INSERT INTO banned_users (id, user_id) VALUES (?, ?) ON CONFLICT(user_id) DO NOTHING", arg.ID, arg.UserID)
	return err
}

func UnbanUser(ctx context.Context, userID string) error {
	_, err := db.BunDB.NewDelete().TableExpr("banned_users").Where("user_id = ?", userID).Exec(ctx)
	return err
}

func IsUserBanned(ctx context.Context, userID string) (int64, error) {
	n, err := db.BunDB.NewSelect().TableExpr("banned_users").Where("user_id = ?", userID).Count(ctx)
	return int64(n), err
}

func CreateUser(ctx context.Context, arg CreateUserParams) (db.User, error) {
	lang := "hu"
	if s, ok := arg.PreferredLang.(string); ok && strings.TrimSpace(s) != "" {
		lang = strings.TrimSpace(s)
	}
	user := db.User{ID: arg.ID, Email: arg.Email, PreferredLang: lang}
	if _, err := db.BunDB.NewInsert().Model(&user).Exec(ctx); err != nil {
		return db.User{}, err
	}
	return GetUserByID(ctx, arg.ID)
}

func GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	var row db.User
	err := db.BunDB.NewSelect().Model(&row).Where("email = ?", email).Scan(ctx)
	return row, err
}

func UpdateUserPreferredLang(ctx context.Context, arg UpdateUserPreferredLangParams) error {
	_, err := db.BunDB.NewUpdate().Model((*db.User)(nil)).
		Set("preferred_lang = ?", arg.PreferredLang).
		Where("id = ?", arg.ID).
		Exec(ctx)
	return err
}

func CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (db.UserSession, error) {
	session := db.UserSession{ID: arg.ID, UserID: arg.UserID, Token: arg.Token, ExpiresAt: arg.ExpiresAt}
	if _, err := db.BunDB.NewInsert().Model(&session).Exec(ctx); err != nil {
		return db.UserSession{}, err
	}
	var row db.UserSession
	err := db.BunDB.NewSelect().Model(&row).Where("id = ?", arg.ID).Scan(ctx)
	return row, err
}

func GetUserSessionByToken(ctx context.Context, token string) (db.UserSession, error) {
	var row db.UserSession
	err := db.BunDB.NewSelect().
		Model(&row).
		Where("token = ?", token).
		Where("expires_at > CURRENT_TIMESTAMP").
		Scan(ctx)
	return row, err
}

func DeleteUserSession(ctx context.Context, arg DeleteUserSessionParams) error {
	_, err := db.BunDB.NewDelete().
		TableExpr("user_sessions").
		Where("id = ?", arg.ID).
		Where("user_id = ?", arg.UserID).
		Exec(ctx)
	return err
}

func DeleteAllUserSessions(ctx context.Context, userID string) error {
	_, err := db.BunDB.NewDelete().TableExpr("user_sessions").Where("user_id = ?", userID).Exec(ctx)
	return err
}

func ListUserSessions(ctx context.Context, userID string) ([]db.UserSession, error) {
	rows := make([]db.UserSession, 0)
	err := db.BunDB.NewSelect().
		Model(&rows).
		Where("user_id = ?", userID).
		OrderExpr("created_at DESC").
		Scan(ctx)
	return rows, err
}

func ListUserDetailCardStates(ctx context.Context, userID string) ([]ListUserDetailCardStatesRow, error) {
	rows := make([]ListUserDetailCardStatesRow, 0)
	err := db.BunDB.NewSelect().
		TableExpr("user_detail_card_states").
		Column("state_key", "is_open").
		Where("user_id = ?", userID).
		Scan(ctx, &rows)
	return rows, err
}

func UpsertUserDetailCardState(ctx context.Context, arg UpsertUserDetailCardStateParams) error {
	_, err := db.BunDB.ExecContext(
		ctx,
		"INSERT INTO user_detail_card_states (user_id, state_key, is_open, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(user_id, state_key) DO UPDATE SET is_open = excluded.is_open, updated_at = CURRENT_TIMESTAMP",
		arg.UserID,
		arg.StateKey,
		arg.IsOpen,
	)
	return err
}

func GetMagicLinkByToken(ctx context.Context, token string) (db.MagicLink, error) {
	var row db.MagicLink
	err := db.BunDB.NewSelect().Model(&row).Where("token = ?", token).Scan(ctx)
	return row, err
}

func UseMagicLink(ctx context.Context, id string) error {
	_, err := db.BunDB.NewUpdate().
		TableExpr("magic_links").
		Set("used_at = CURRENT_TIMESTAMP").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func CreateMagicLink(ctx context.Context, arg CreateMagicLinkParams) (db.MagicLink, error) {
	var row db.MagicLink
	err := db.BunDB.QueryRowContext(
		ctx,
		"INSERT INTO magic_links (id, token, email, action, group_id, expires_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id, token, email, action, group_id, expires_at, used_at, created_at, invite_role",
		arg.ID,
		arg.Token,
		arg.Email,
		arg.Action,
		arg.GroupID,
		arg.ExpiresAt,
	).Scan(&row.ID, &row.Token, &row.Email, &row.Action, &row.GroupID, &row.ExpiresAt, &row.UsedAt, &row.CreatedAt, &row.InviteRole)
	return row, err
}
