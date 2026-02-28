-- name: CreateExpense :one
INSERT INTO expenses (id, group_id, title, description, amount, date)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: ListExpenses :many
SELECT * FROM expenses
WHERE group_id = ?
ORDER BY date DESC, created_at DESC;

-- name: CountExpensesFiltered :one
SELECT COUNT(*) FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListExpensesByDateAscFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY date ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByDateDescFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY date DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByTitleAscFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY title COLLATE NOCASE ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByTitleDescFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY title COLLATE NOCASE DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByAmountAscFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY amount ASC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByAmountDescFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY amount DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateExpense :one
UPDATE expenses
SET title = ?, description = ?, amount = ?, date = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = ? AND group_id = ?;
