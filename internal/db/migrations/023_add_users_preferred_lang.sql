-- +goose Up
ALTER TABLE users ADD COLUMN preferred_lang TEXT NOT NULL DEFAULT 'hu';

-- +goose Down
ALTER TABLE users DROP COLUMN preferred_lang;
