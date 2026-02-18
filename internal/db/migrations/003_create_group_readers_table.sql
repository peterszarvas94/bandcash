-- +goose Up
CREATE TABLE IF NOT EXISTS group_readers (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE(user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_readers_user_id ON group_readers(user_id);
CREATE INDEX IF NOT EXISTS idx_group_readers_group_id ON group_readers(group_id);

-- +goose Down
DROP INDEX IF EXISTS idx_group_readers_group_id;
DROP INDEX IF EXISTS idx_group_readers_user_id;
DROP TABLE IF EXISTS group_readers;
