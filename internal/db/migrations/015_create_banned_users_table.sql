-- +goose Up
CREATE TABLE IF NOT EXISTS banned_users (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL UNIQUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_banned_users_user_id ON banned_users(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_banned_users_user_id;
DROP TABLE IF EXISTS banned_users;
