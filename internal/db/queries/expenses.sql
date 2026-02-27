-- name: CreateExpense :one
INSERT INTO expenses (id, group_id, title, description, amount, date)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListExpenses :many
SELECT * FROM expenses
WHERE group_id = ?
ORDER BY date DESC, created_at DESC;

-- name: UpdateExpense :one
UPDATE expenses
SET title = ?, description = ?, amount = ?, date = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = ? AND group_id = ?;
