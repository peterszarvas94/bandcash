-- name: CreateExpense :one
INSERT INTO expenses (id, group_id, title, description, amount, date, paid)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetExpense :one
SELECT * FROM expenses
WHERE id = ? AND group_id = ?;

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
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  );

-- name: ListExpensesByDateAscFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
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
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
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
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
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
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
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
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
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
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY amount DESC, created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByDescriptionAscFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY description COLLATE NOCASE ASC, date DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListExpensesByDescriptionDescFiltered :many
SELECT * FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY description COLLATE NOCASE DESC, date DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: UpdateExpense :one
UPDATE expenses
SET title = ?, description = ?, amount = ?, date = ?, paid = ?
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: ToggleExpensePaid :one
UPDATE expenses
SET paid = CASE WHEN paid = 1 THEN 0 ELSE 1 END
WHERE id = ? AND group_id = ?
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = ? AND group_id = ?;

-- name: SumExpensesFiltered :one
SELECT CAST(COALESCE(SUM(amount), 0) AS INTEGER) FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  );

-- name: SumExpenseTotalsFiltered :one
SELECT
  CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS total_paid,
  CAST(COALESCE(SUM(CASE WHEN paid = 0 THEN amount ELSE 0 END), 0) AS INTEGER) AS total_unpaid
FROM expenses
WHERE group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR title LIKE '%' || sqlc.arg(search) || '%'
    OR description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(expenses.date) >= date(sqlc.arg(from_date))
      AND date(expenses.date) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', expenses.date) = sqlc.arg(year_filter)
      )
    )
  );
