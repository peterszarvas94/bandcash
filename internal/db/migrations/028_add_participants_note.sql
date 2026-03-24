-- +goose Up
ALTER TABLE participants ADD COLUMN note TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE participants DROP COLUMN note;
