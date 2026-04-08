-- +goose Up
CREATE VIEW IF NOT EXISTS group_outgoing_payments AS
SELECT
  p.group_id AS group_id,
  'participant' AS payment_kind,
  CAST(p.event_id || ':' || p.member_id AS TEXT) AS payment_id,
  CAST(p.event_id AS TEXT) AS event_id,
  CAST(p.member_id AS TEXT) AS member_id,
  CAST(m.name AS TEXT) AS member_name,
  CAST(e.title AS TEXT) AS event_title,
  e.title AS title,
  CAST((p.amount + p.expense) AS INTEGER) AS amount,
  p.paid AS paid,
  p.paid_at AS paid_at,
  p.updated_at AS updated_at,
  e.time AS sort_date
FROM participants p
JOIN members m ON m.id = p.member_id AND m.group_id = p.group_id
JOIN events e ON e.id = p.event_id AND e.group_id = p.group_id
UNION ALL
SELECT
  ex.group_id AS group_id,
  'expense' AS payment_kind,
  CAST(ex.id AS TEXT) AS payment_id,
  '' AS event_id,
  '' AS member_id,
  '' AS member_name,
  '' AS event_title,
  ex.title AS title,
  CAST(ex.amount AS INTEGER) AS amount,
  ex.paid AS paid,
  ex.paid_at AS paid_at,
  ex.updated_at AS updated_at,
  ex.date AS sort_date
FROM expenses ex;

-- +goose Down
DROP VIEW IF EXISTS group_outgoing_payments;
