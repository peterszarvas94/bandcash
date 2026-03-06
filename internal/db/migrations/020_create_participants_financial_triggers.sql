-- +goose Up
-- +goose StatementBegin

-- Trigger: After INSERT on participants - add amount to group totals
CREATE TRIGGER IF NOT EXISTS trg_participants_insert_financials
AFTER INSERT ON participants
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_payout_amount = total_payout_amount + NEW.amount,
        total_leftover = total_leftover - NEW.amount
    WHERE id = NEW.group_id;
END;

-- Trigger: After UPDATE on participants - adjust amount difference
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

-- Trigger: After DELETE on participants - subtract amount from group totals
CREATE TRIGGER IF NOT EXISTS trg_participants_delete_financials
AFTER DELETE ON participants
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_payout_amount = total_payout_amount - OLD.amount,
        total_leftover = total_leftover + OLD.amount
    WHERE id = OLD.group_id;
END;

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_participants_insert_financials;
DROP TRIGGER IF EXISTS trg_participants_update_financials;
DROP TRIGGER IF EXISTS trg_participants_delete_financials;
