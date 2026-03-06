-- +goose Up
ALTER TABLE groups ADD COLUMN total_event_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_expense_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_payout_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_leftover INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE groups DROP COLUMN total_event_amount;
ALTER TABLE groups DROP COLUMN total_expense_amount;
ALTER TABLE groups DROP COLUMN total_payout_amount;
ALTER TABLE groups DROP COLUMN total_leftover;
