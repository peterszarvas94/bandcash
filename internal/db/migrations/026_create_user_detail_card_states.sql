-- +goose Up
CREATE TABLE user_detail_card_states (
  user_id TEXT NOT NULL,
  state_key TEXT NOT NULL,
  is_open INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, state_key),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_detail_card_states_user_id ON user_detail_card_states(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_detail_card_states_user_id;
DROP TABLE IF EXISTS user_detail_card_states;
