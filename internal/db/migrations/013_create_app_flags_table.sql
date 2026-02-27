-- +goose Up
CREATE TABLE IF NOT EXISTS app_flags (
  key TEXT PRIMARY KEY,
  bool_value INTEGER NOT NULL CHECK (bool_value IN (0, 1)),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO app_flags (key, bool_value)
VALUES ('enable_signup', 0)
ON CONFLICT(key) DO NOTHING;

-- +goose Down
DELETE FROM app_flags WHERE key = 'enable_signup';
DROP TABLE IF EXISTS app_flags;
