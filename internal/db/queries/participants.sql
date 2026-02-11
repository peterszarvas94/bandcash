-- name: AddParticipant :one
INSERT INTO participants (entry_id, payee_id, amount)
VALUES (?, ?, ?)
RETURNING *;

-- name: RemoveParticipant :exec
DELETE FROM participants
WHERE entry_id = ? AND payee_id = ?;

-- name: ListParticipantsByEntry :many
SELECT payees.*, participants.amount AS participant_amount
FROM payees
JOIN participants ON participants.payee_id = payees.id
WHERE participants.entry_id = ?
ORDER BY payees.name ASC;

-- name: ListParticipantsByPayee :many
SELECT entries.*, participants.amount AS participant_amount
FROM entries
JOIN participants ON participants.entry_id = entries.id
WHERE participants.payee_id = ?
ORDER BY entries.created_at DESC;
