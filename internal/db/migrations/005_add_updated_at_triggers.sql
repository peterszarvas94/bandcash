-- +goose Up
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

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS participants_updated_at;
DROP TRIGGER IF EXISTS payees_updated_at;
DROP TRIGGER IF EXISTS entries_updated_at;
-- +goose StatementEnd
