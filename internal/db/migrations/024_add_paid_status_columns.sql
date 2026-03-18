-- +goose Up
-- Add paid status columns to events, expenses, and participants tables
ALTER TABLE events ADD COLUMN paid INTEGER NOT NULL DEFAULT 0;
ALTER TABLE expenses ADD COLUMN paid INTEGER NOT NULL DEFAULT 0;
ALTER TABLE participants ADD COLUMN paid INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- Remove paid status columns
ALTER TABLE events DROP COLUMN paid;
ALTER TABLE expenses DROP COLUMN paid;
ALTER TABLE participants DROP COLUMN paid;
