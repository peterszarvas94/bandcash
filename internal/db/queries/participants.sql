-- name: AddParticipant :one
INSERT INTO participants (event_id, member_id, amount, expense)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: RemoveParticipant :exec
DELETE FROM participants
WHERE event_id = ? AND member_id = ?;

-- name: UpdateParticipant :exec
UPDATE participants
SET amount = ?, expense = ?
WHERE event_id = ? AND member_id = ?;

-- name: ListParticipantsByEvent :many
SELECT members.*, participants.amount AS participant_amount, participants.expense AS participant_expense
FROM members
JOIN participants ON participants.member_id = members.id
WHERE participants.event_id = ?
ORDER BY members.name ASC;

-- name: ListParticipantsByMember :many
SELECT events.*, participants.amount AS participant_amount, participants.expense AS participant_expense
FROM events
JOIN participants ON participants.event_id = events.id
WHERE participants.member_id = ?
ORDER BY events.created_at DESC;
