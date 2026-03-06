-- +goose Up
-- +goose StatementBegin

-- Trigger: After INSERT on expenses - add amount to group totals
CREATE TRIGGER IF NOT EXISTS trg_expenses_insert_financials
AFTER INSERT ON expenses
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_expense_amount = total_expense_amount + NEW.amount,
        total_leftover = total_leftover - NEW.amount
    WHERE id = NEW.group_id;
END;

-- Trigger: After UPDATE on expenses - adjust amount difference
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

-- Trigger: After DELETE on expenses - subtract amount from group totals
CREATE TRIGGER IF NOT EXISTS trg_expenses_delete_financials
AFTER DELETE ON expenses
FOR EACH ROW
BEGIN
    UPDATE groups 
    SET total_expense_amount = total_expense_amount - OLD.amount,
        total_leftover = total_leftover + OLD.amount
    WHERE id = OLD.group_id;
END;

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_expenses_insert_financials;
DROP TRIGGER IF EXISTS trg_expenses_update_financials;
DROP TRIGGER IF EXISTS trg_expenses_delete_financials;
