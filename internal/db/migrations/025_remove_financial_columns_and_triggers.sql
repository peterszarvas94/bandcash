-- +goose Up
-- Production-safe removal of financial calculation columns and triggers
-- These are now calculated in-memory and cached

-- Drop triggers first (SQLite requires this before dropping columns they reference)
DROP TRIGGER IF EXISTS trg_events_insert_financials;
DROP TRIGGER IF EXISTS trg_events_update_financials;
DROP TRIGGER IF EXISTS trg_events_delete_financials;
DROP TRIGGER IF EXISTS trg_expenses_insert_financials;
DROP TRIGGER IF EXISTS trg_expenses_update_financials;
DROP TRIGGER IF EXISTS trg_expenses_delete_financials;
DROP TRIGGER IF EXISTS trg_participants_insert_financials;
DROP TRIGGER IF EXISTS trg_participants_update_financials;
DROP TRIGGER IF EXISTS trg_participants_delete_financials;

-- Drop financial columns from groups table
ALTER TABLE groups DROP COLUMN total_event_amount;
ALTER TABLE groups DROP COLUMN total_expense_amount;
ALTER TABLE groups DROP COLUMN total_payout_amount;
ALTER TABLE groups DROP COLUMN total_leftover;

-- +goose Down
-- Restore financial columns (will be recalculated on next app start)
ALTER TABLE groups ADD COLUMN total_event_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_expense_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_payout_amount INTEGER NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN total_leftover INTEGER NOT NULL DEFAULT 0;

-- Recreate triggers (simplified without paid status logic for down migration)
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_events_insert_financials
AFTER INSERT ON events
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_event_amount = total_event_amount + NEW.amount,
        total_leftover = total_leftover + NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_events_update_financials
AFTER UPDATE ON events
FOR EACH ROW
WHEN OLD.amount != NEW.amount
BEGIN
    UPDATE groups 
    SET total_event_amount = total_event_amount - OLD.amount + NEW.amount,
        total_leftover = total_leftover - OLD.amount + NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_events_delete_financials
AFTER DELETE ON events
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_event_amount = total_event_amount - OLD.amount,
        total_leftover = total_leftover - OLD.amount
    WHERE id = OLD.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_expenses_insert_financials
AFTER INSERT ON expenses
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_expense_amount = total_expense_amount + NEW.amount,
        total_leftover = total_leftover - NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_expenses_update_financials
AFTER UPDATE ON expenses
FOR EACH ROW
WHEN OLD.amount != NEW.amount
BEGIN
    UPDATE groups 
    SET total_expense_amount = total_expense_amount - OLD.amount + NEW.amount,
        total_leftover = total_leftover + OLD.amount - NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_expenses_delete_financials
AFTER DELETE ON expenses
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_expense_amount = total_expense_amount - OLD.amount,
        total_leftover = total_leftover + OLD.amount
    WHERE id = OLD.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_participants_insert_financials
AFTER INSERT ON participants
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_payout_amount = total_payout_amount + NEW.amount,
        total_leftover = total_leftover - NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_participants_update_financials
AFTER UPDATE ON participants
FOR EACH ROW
WHEN OLD.amount != NEW.amount
BEGIN
    UPDATE groups 
    SET total_payout_amount = total_payout_amount - OLD.amount + NEW.amount,
        total_leftover = total_leftover + OLD.amount - NEW.amount
    WHERE id = NEW.group_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_participants_delete_financials
AFTER DELETE ON participants
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_payout_amount = total_payout_amount - OLD.amount,
        total_leftover = total_leftover + NEW.amount
    WHERE id = OLD.group_id;
END;
-- +goose StatementEnd
