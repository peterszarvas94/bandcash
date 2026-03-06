-- +goose Up
-- +goose StatementBegin

-- Trigger: After INSERT on events - add amount to group totals
CREATE TRIGGER IF NOT EXISTS trg_events_insert_financials
AFTER INSERT ON events
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_event_amount = total_event_amount + NEW.amount,
        total_leftover = total_leftover + NEW.amount
    WHERE id = NEW.group_id;
END;

-- Trigger: After UPDATE on events - adjust amount difference
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

-- Trigger: After DELETE on events - subtract amount from group totals
CREATE TRIGGER IF NOT EXISTS trg_events_delete_financials
AFTER DELETE ON events
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_event_amount = total_event_amount - OLD.amount,
        total_leftover = total_leftover - OLD.amount
    WHERE id = OLD.group_id;
END;

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_events_insert_financials;
DROP TRIGGER IF EXISTS trg_events_update_financials;
DROP TRIGGER IF EXISTS trg_events_delete_financials;
