-- name: CountGroupPendingPaymentRowsFiltered :one
SELECT COUNT(*) FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  );

-- name: ListGroupPendingPaymentRowsByDateAscFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY date_value ASC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingPaymentRowsByDateDescFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY date_value DESC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingPaymentRowsByNameAscFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY name COLLATE NOCASE ASC, date_value DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingPaymentRowsByNameDescFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY name COLLATE NOCASE DESC, date_value DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingPaymentRowsByAmountAscFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY amount ASC, date_value DESC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListGroupPendingPaymentRowsByAmountDescFiltered :many
SELECT * FROM group_payment_rows
WHERE group_id = sqlc.arg(group_id)
  AND paid = 0
  AND (
    sqlc.arg(search) = ''
    OR name LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    (
      sqlc.arg(from_date) != ''
      AND sqlc.arg(to_date) != ''
      AND date(date_value) >= date(sqlc.arg(from_date))
      AND date(date_value) <= date(sqlc.arg(to_date))
    )
    OR (
      (sqlc.arg(from_date) = '' OR sqlc.arg(to_date) = '')
      AND (
        sqlc.arg(year_filter) = ''
        OR strftime('%Y', date_value) = sqlc.arg(year_filter)
      )
    )
  )
ORDER BY amount DESC, date_value DESC, name COLLATE NOCASE ASC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);
