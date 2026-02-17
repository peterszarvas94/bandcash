-- +goose Up
-- +goose StatementBegin
DROP TRIGGER IF EXISTS participants_updated_at;
DROP TRIGGER IF EXISTS payees_updated_at;
DROP TRIGGER IF EXISTS entries_updated_at;
-- +goose StatementEnd

ALTER TABLE entries RENAME TO events;
ALTER TABLE payees RENAME TO members;

ALTER TABLE participants RENAME COLUMN entry_id TO event_id;
ALTER TABLE participants RENAME COLUMN payee_id TO member_id;

DROP INDEX IF EXISTS participants_entry_id_idx;
DROP INDEX IF EXISTS participants_payee_id_idx;

CREATE INDEX IF NOT EXISTS participants_event_id_idx ON participants(event_id);
CREATE INDEX IF NOT EXISTS participants_member_id_idx ON participants(member_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS events_updated_at
AFTER UPDATE ON events
FOR EACH ROW
BEGIN
    UPDATE events SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS members_updated_at
AFTER UPDATE ON members
FOR EACH ROW
BEGIN
    UPDATE members SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS participants_updated_at
AFTER UPDATE ON participants
FOR EACH ROW
BEGIN
    UPDATE participants
    SET updated_at = CURRENT_TIMESTAMP
    WHERE event_id = NEW.event_id AND member_id = NEW.member_id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS participants_updated_at;
DROP TRIGGER IF EXISTS members_updated_at;
DROP TRIGGER IF EXISTS events_updated_at;
-- +goose StatementEnd

DROP INDEX IF EXISTS participants_member_id_idx;
DROP INDEX IF EXISTS participants_event_id_idx;

ALTER TABLE participants RENAME COLUMN member_id TO payee_id;
ALTER TABLE participants RENAME COLUMN event_id TO entry_id;

ALTER TABLE members RENAME TO payees;
ALTER TABLE events RENAME TO entries;

CREATE INDEX IF NOT EXISTS participants_entry_id_idx ON participants(entry_id);
CREATE INDEX IF NOT EXISTS participants_payee_id_idx ON participants(payee_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS entries_updated_at
AFTER UPDATE ON entries
FOR EACH ROW
BEGIN
    UPDATE entries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS payees_updated_at
AFTER UPDATE ON payees
FOR EACH ROW
BEGIN
    UPDATE payees SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS participants_updated_at
AFTER UPDATE ON participants
FOR EACH ROW
BEGIN
    UPDATE participants
    SET updated_at = CURRENT_TIMESTAMP
    WHERE entry_id = NEW.entry_id AND payee_id = NEW.payee_id;
END;
-- +goose StatementEnd
