// source: settings.sql

package db

import (
	"context"
)

const listUserDetailCardStates = `-- name: ListUserDetailCardStates :many
SELECT state_key, is_open
FROM user_detail_card_states
WHERE user_id = ?
`

type ListUserDetailCardStatesRow struct {
	StateKey string `json:"state_key"`
	IsOpen   int64  `json:"is_open"`
}

func (q *Queries) ListUserDetailCardStates(ctx context.Context, userID string) ([]ListUserDetailCardStatesRow, error) {
	rows, err := q.db.QueryContext(ctx, listUserDetailCardStates, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUserDetailCardStatesRow{}
	for rows.Next() {
		var i ListUserDetailCardStatesRow
		if err := rows.Scan(&i.StateKey, &i.IsOpen); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertUserDetailCardState = `-- name: UpsertUserDetailCardState :exec
INSERT INTO user_detail_card_states (user_id, state_key, is_open, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, state_key) DO UPDATE SET
  is_open = excluded.is_open,
  updated_at = CURRENT_TIMESTAMP
`

type UpsertUserDetailCardStateParams struct {
	UserID   string `json:"user_id"`
	StateKey string `json:"state_key"`
	IsOpen   int64  `json:"is_open"`
}

func (q *Queries) UpsertUserDetailCardState(ctx context.Context, arg UpsertUserDetailCardStateParams) error {
	_, err := q.db.ExecContext(ctx, upsertUserDetailCardState, arg.UserID, arg.StateKey, arg.IsOpen)
	return err
}
