-- +goose Up
-- Backfill cached financial totals for all existing groups

UPDATE groups SET
    total_event_amount = COALESCE((SELECT SUM(amount) FROM events WHERE events.group_id = groups.id), 0),
    total_expense_amount = COALESCE((SELECT SUM(amount) FROM expenses WHERE expenses.group_id = groups.id), 0),
    total_payout_amount = COALESCE((SELECT SUM(amount) FROM participants WHERE participants.group_id = groups.id), 0),
    total_leftover = COALESCE((SELECT SUM(amount) FROM events WHERE events.group_id = groups.id), 0) 
                     - COALESCE((SELECT SUM(amount) FROM expenses WHERE expenses.group_id = groups.id), 0)
                     - COALESCE((SELECT SUM(amount) FROM participants WHERE participants.group_id = groups.id), 0);

-- +goose Down
-- Reset all cached values to 0
UPDATE groups SET
    total_event_amount = 0,
    total_expense_amount = 0,
    total_payout_amount = 0,
    total_leftover = 0;
