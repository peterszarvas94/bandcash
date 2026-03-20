-- name: UpsertUserDetailCardState :exec
INSERT INTO user_detail_card_states (user_id, state_key, is_open, updated_at)
VALUES (?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(user_id, state_key) DO UPDATE SET
  is_open = excluded.is_open,
  updated_at = CURRENT_TIMESTAMP;

-- name: ListUserDetailCardStates :many
SELECT state_key, is_open
FROM user_detail_card_states
WHERE user_id = ?;
