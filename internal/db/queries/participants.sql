-- name: AddParticipant :one
INSERT INTO participants (group_id, event_id, member_id, amount, expense, note, paid, paid_at)
VALUES (
  sqlc.arg(group_id),
  sqlc.arg(event_id),
  sqlc.arg(member_id),
  sqlc.arg(amount),
  sqlc.arg(expense),
  sqlc.arg(note),
  sqlc.arg(paid),
  CASE
    WHEN sqlc.arg(paid) = 1 THEN COALESCE(sqlc.narg(paid_at), CURRENT_TIMESTAMP)
    ELSE NULL
  END
)
RETURNING *;

-- name: RemoveParticipant :exec
DELETE FROM participants
WHERE event_id = ? AND member_id = ? AND group_id = ?;

-- name: UpdateParticipant :exec
UPDATE participants
SET amount = sqlc.arg(amount),
    expense = sqlc.arg(expense),
    note = sqlc.arg(note),
    paid = sqlc.arg(paid),
    paid_at = CASE
      WHEN sqlc.arg(paid) = 0 THEN NULL
      WHEN sqlc.narg(paid_at) IS NOT NULL THEN sqlc.narg(paid_at)
      WHEN paid = 0 THEN CURRENT_TIMESTAMP
      ELSE paid_at
    END
WHERE event_id = sqlc.arg(event_id)
  AND member_id = sqlc.arg(member_id)
  AND group_id = sqlc.arg(group_id);

-- name: ToggleParticipantPaid :one
UPDATE participants
SET paid = CASE WHEN paid = 1 THEN 0 ELSE 1 END,
    paid_at = CASE WHEN paid = 1 THEN NULL ELSE CURRENT_TIMESTAMP END
WHERE event_id = ? AND member_id = ? AND group_id = ?
RETURNING *;

-- name: UpdateParticipantNote :exec
UPDATE participants
SET note = sqlc.arg(note)
WHERE event_id = sqlc.arg(event_id)
  AND member_id = sqlc.arg(member_id)
  AND group_id = sqlc.arg(group_id);

-- name: UpdateParticipantPaidAt :exec
UPDATE participants
SET paid = 1,
    paid_at = NULLIF(sqlc.narg(paid_at), '')
WHERE event_id = sqlc.arg(event_id)
  AND member_id = sqlc.arg(member_id)
  AND group_id = sqlc.arg(group_id);

-- name: ListParticipantsByEvent :many
SELECT members.*, participants.amount AS participant_amount, participants.expense AS participant_expense, participants.note AS participant_note, participants.paid AS participant_paid, participants.paid_at AS participant_paid_at
FROM members
JOIN participants ON participants.member_id = members.id
WHERE participants.event_id = ? AND participants.group_id = ?
ORDER BY members.name ASC;

-- name: ListParticipantsByMember :many
SELECT events.*, participants.amount AS participant_amount, participants.expense AS participant_expense, participants.paid AS participant_paid, participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = ? AND participants.group_id = ?
ORDER BY events.created_at DESC;

-- name: ListRecentPaidParticipantsByGroup :many
SELECT
  participants.event_id,
  participants.member_id,
  participants.amount AS participant_amount,
  participants.expense AS participant_expense,
  participants.paid_at AS participant_paid_at,
  participants.updated_at AS participant_updated_at,
  members.name AS member_name
FROM participants
JOIN members ON members.id = participants.member_id AND members.group_id = participants.group_id
WHERE participants.group_id = sqlc.arg(group_id)
  AND participants.paid = 1
ORDER BY participants.updated_at DESC
LIMIT sqlc.arg(limit);

-- name: SumParticipantAmountsByGroup :one
SELECT CAST(COALESCE(SUM(amount), 0) AS INTEGER) FROM participants
WHERE group_id = ?;

-- name: SumParticipantPaidAmountsByGroup :one
SELECT
  CAST(COALESCE(SUM(CASE WHEN paid = 1 THEN amount ELSE 0 END), 0) AS INTEGER) AS paid_amount,
  CAST(COALESCE(SUM(CASE WHEN paid = 0 THEN amount ELSE 0 END), 0) AS INTEGER) AS unpaid_amount
FROM participants
WHERE group_id = ?;

-- name: SumParticipantTotalsByGroupFiltered :one
SELECT
  CAST(COALESCE(SUM(CASE WHEN participants.paid = 1 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_paid,
  CAST(COALESCE(SUM(CASE WHEN participants.paid = 0 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) AS total_unpaid
FROM participants
JOIN events ON events.id = participants.event_id
WHERE participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  );

-- name: CountParticipantsByMemberFiltered :one
SELECT COUNT(*) FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  );

-- name: ListParticipantsByMemberByTitleAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.title COLLATE NOCASE ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTitleDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.title COLLATE NOCASE DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTimeAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.time ASC, events.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTimeDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.time DESC, events.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByAmountAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.amount ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByAmountDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY events.amount DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByCutAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.amount ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByCutDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.amount DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByExpenseAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.expense ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByExpenseDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.expense DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByPaidAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.paid ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByPaidDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.paid DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByPaidAtAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.paid_at ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByPaidAtDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid,
  participants.paid_at AS participant_paid_at
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  )
ORDER BY participants.paid_at DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SumParticipantTotalsByMemberFiltered :one
SELECT 
  CAST(COALESCE(SUM(participants.amount), 0) AS INTEGER) as total_cut,
  CAST(COALESCE(SUM(participants.expense), 0) AS INTEGER) as total_expense,
  CAST(COALESCE(SUM(participants.amount + participants.expense), 0) AS INTEGER) as total_payout,
  CAST(COALESCE(SUM(CASE WHEN participants.paid = 1 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) as total_paid,
  CAST(COALESCE(SUM(CASE WHEN participants.paid = 0 THEN participants.amount + participants.expense ELSE 0 END), 0) AS INTEGER) as total_unpaid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
  AND (
    sqlc.arg(year) = ''
    OR events.time LIKE sqlc.arg(year) || '%'
  )
  AND (
    sqlc.arg(from) = ''
    OR events.time >= sqlc.arg(from)
  )
  AND (
    sqlc.arg(to) = ''
    OR events.time <= sqlc.arg(to)
  );
