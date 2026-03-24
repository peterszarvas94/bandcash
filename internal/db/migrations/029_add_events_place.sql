-- +goose Up
ALTER TABLE events ADD COLUMN place TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE events DROP COLUMN place;
