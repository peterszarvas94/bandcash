-- +goose Up
CREATE TABLE IF NOT EXISTS participants (
    event_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    expense INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, member_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS participants;
