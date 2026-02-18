-- +goose Up
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_participants_updated_at
AFTER UPDATE ON participants
FOR EACH ROW
BEGIN
    UPDATE participants
    SET updated_at = CURRENT_TIMESTAMP
    WHERE event_id = NEW.event_id AND member_id = NEW.member_id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_participants_updated_at;
