-- +goose Up
ALTER TABLE events ADD COLUMN paid_at TEXT;
ALTER TABLE expenses ADD COLUMN paid_at TEXT;
ALTER TABLE participants ADD COLUMN paid_at TEXT;

-- +goose Down
ALTER TABLE events DROP COLUMN paid_at;
ALTER TABLE expenses DROP COLUMN paid_at;
ALTER TABLE participants DROP COLUMN paid_at;
