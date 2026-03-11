-- +goose Up
CREATE TABLE IF NOT EXISTS group_admins (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE(user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_admins_user_id ON group_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_group_admins_group_id ON group_admins(group_id);

ALTER TABLE magic_links
ADD COLUMN invite_role TEXT NOT NULL DEFAULT 'viewer' CHECK (invite_role IN ('viewer', 'admin'));

UPDATE magic_links
SET invite_role = 'viewer'
WHERE action = 'invite';

-- +goose Down
ALTER TABLE magic_links DROP COLUMN invite_role;

DROP INDEX IF EXISTS idx_group_admins_group_id;
DROP INDEX IF EXISTS idx_group_admins_user_id;
DROP TABLE IF EXISTS group_admins;
