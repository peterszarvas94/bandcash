-- name: CreatePayee :one
INSERT INTO payees (name, description)
VALUES (?, ?)
RETURNING *;

-- name: GetPayee :one
SELECT * FROM payees
WHERE id = ?;

-- name: ListPayees :many
SELECT * FROM payees
ORDER BY created_at DESC;

-- name: UpdatePayee :one
UPDATE payees
SET name = ?, description = ?
WHERE id = ?
RETURNING *;

-- name: DeletePayee :exec
DELETE FROM payees
WHERE id = ?;
