-- name: AddParticipant :one
INSERT INTO participants (group_id, event_id, member_id, amount, expense, paid)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: RemoveParticipant :exec
DELETE FROM participants
WHERE event_id = ? AND member_id = ? AND group_id = ?;

-- name: UpdateParticipant :exec
UPDATE participants
SET amount = ?, expense = ?
WHERE event_id = ? AND member_id = ? AND group_id = ?;

-- name: ToggleParticipantPaid :one
UPDATE participants
SET paid = CASE WHEN paid = 1 THEN 0 ELSE 1 END
WHERE event_id = ? AND member_id = ? AND group_id = ?
RETURNING *;

-- name: ListParticipantsByEvent :many
SELECT members.*, participants.amount AS participant_amount, participants.expense AS participant_expense, participants.paid AS participant_paid
FROM members
JOIN participants ON participants.member_id = members.id
WHERE participants.event_id = ? AND participants.group_id = ?
ORDER BY members.name ASC;

-- name: ListParticipantsByMember :many
SELECT events.*, participants.amount AS participant_amount, participants.expense AS participant_expense, participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = ? AND participants.group_id = ?
ORDER BY events.created_at DESC;

-- name: SumParticipantAmountsByGroup :one
SELECT CAST(COALESCE(SUM(amount), 0) AS INTEGER) FROM participants
WHERE group_id = ?;

-- name: CountParticipantsByMemberFiltered :one
SELECT COUNT(*) FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  );

-- name: ListParticipantsByMemberByTitleAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.title COLLATE NOCASE ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTitleDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.title COLLATE NOCASE DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTimeAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.time ASC, events.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByTimeDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.time DESC, events.created_at DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByAmountAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.amount ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByAmountDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY events.amount DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByCutAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY participants.amount ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByCutDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY participants.amount DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByExpenseAscFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY participants.expense ASC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: ListParticipantsByMemberByExpenseDescFiltered :many
SELECT 
  events.*, 
  participants.amount AS participant_amount, 
  participants.expense AS participant_expense,
  participants.paid AS participant_paid
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  )
ORDER BY participants.expense DESC, events.time DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: SumParticipantTotalsByMemberFiltered :one
SELECT 
  CAST(COALESCE(SUM(participants.amount), 0) AS INTEGER) as total_cut,
  CAST(COALESCE(SUM(participants.expense), 0) AS INTEGER) as total_expense,
  CAST(COALESCE(SUM(participants.amount + participants.expense), 0) AS INTEGER) as total_payout
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = sqlc.arg(member_id)
  AND participants.group_id = sqlc.arg(group_id)
  AND (
    sqlc.arg(search) = ''
    OR events.title LIKE '%' || sqlc.arg(search) || '%'
    OR events.description LIKE '%' || sqlc.arg(search) || '%'
  );
