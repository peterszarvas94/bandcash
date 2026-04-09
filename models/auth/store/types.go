package store

import (
	"database/sql"
	"time"
)

type CreateUserParams struct {
	ID            string      `json:"id"`
	Email         string      `json:"email"`
	PreferredLang interface{} `json:"preferred_lang"`
}

type CreateUserSessionParams struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DeleteUserSessionParams struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type BanUserParams struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type UpdateUserPreferredLangParams struct {
	PreferredLang string `json:"preferred_lang"`
	ID            string `json:"id"`
}

type UpsertAppFlagBoolParams struct {
	Key       string `json:"key"`
	BoolValue int64  `json:"bool_value"`
}

type ListUserDetailCardStatesRow struct {
	StateKey string `json:"state_key"`
	IsOpen   int64  `json:"is_open"`
}

type UpsertUserDetailCardStateParams struct {
	UserID   string `json:"user_id"`
	StateKey string `json:"state_key"`
	IsOpen   int64  `json:"is_open"`
}

type CreateMagicLinkParams struct {
	ID        string         `json:"id"`
	Token     string         `json:"token"`
	Email     string         `json:"email"`
	Action    string         `json:"action"`
	GroupID   sql.NullString `json:"group_id"`
	ExpiresAt time.Time      `json:"expires_at"`
}
