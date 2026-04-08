-- +goose Up
CREATE VIEW group_payment_rows AS
SELECT
    events.group_id AS group_id,
    'event' AS kind,
    events.id AS entity_id,
    events.id AS event_id,
    '' AS member_id,
    events.title AS name,
    events.time AS date_value,
    events.amount AS amount,
    events.paid AS paid,
    events.paid_at AS paid_at
FROM events

UNION ALL

SELECT
    participants.group_id AS group_id,
    'participant' AS kind,
    participants.event_id || ':' || participants.member_id AS entity_id,
    participants.event_id AS event_id,
    participants.member_id AS member_id,
    members.name || ' (' || events.title || ')' AS name,
    events.time AS date_value,
    participants.amount + participants.expense AS amount,
    participants.paid AS paid,
    participants.paid_at AS paid_at
FROM participants
JOIN members
    ON members.id = participants.member_id
   AND members.group_id = participants.group_id
JOIN events
    ON events.id = participants.event_id
   AND events.group_id = participants.group_id

UNION ALL

SELECT
    expenses.group_id AS group_id,
    'expense' AS kind,
    expenses.id AS entity_id,
    '' AS event_id,
    '' AS member_id,
    expenses.title AS name,
    expenses.date AS date_value,
    expenses.amount AS amount,
    expenses.paid AS paid,
    expenses.paid_at AS paid_at
FROM expenses;

-- +goose Down
DROP VIEW IF EXISTS group_payment_rows;
