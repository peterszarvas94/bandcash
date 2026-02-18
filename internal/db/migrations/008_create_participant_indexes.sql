-- +goose Up
CREATE INDEX IF NOT EXISTS idx_participants_event_id ON participants(event_id);
CREATE INDEX IF NOT EXISTS idx_participants_member_id ON participants(member_id);

-- +goose Down
DROP INDEX IF EXISTS idx_participants_member_id;
DROP INDEX IF EXISTS idx_participants_event_id;
