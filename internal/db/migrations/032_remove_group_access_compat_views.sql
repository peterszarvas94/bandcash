-- +goose Up
DROP TRIGGER IF EXISTS trg_group_readers_view_delete;
DROP TRIGGER IF EXISTS trg_group_readers_view_insert;
DROP TRIGGER IF EXISTS trg_group_admins_view_delete;
DROP TRIGGER IF EXISTS trg_group_admins_view_insert;
DROP VIEW IF EXISTS group_readers;
DROP VIEW IF EXISTS group_admins;

-- +goose Down
CREATE VIEW group_admins AS
SELECT id, user_id, group_id, created_at
FROM group_access
WHERE role = 'admin';

CREATE VIEW group_readers AS
SELECT id, user_id, group_id, created_at
FROM group_access
WHERE role = 'viewer';

-- +goose StatementBegin
CREATE TRIGGER trg_group_admins_view_insert
INSTEAD OF INSERT ON group_admins
BEGIN
    INSERT INTO group_access (id, user_id, group_id, role, created_at)
    VALUES (
        COALESCE(NEW.id, 'gac_' || lower(hex(randomblob(10)))),
        NEW.user_id,
        NEW.group_id,
        'admin',
        COALESCE(NEW.created_at, CURRENT_TIMESTAMP)
    )
    ON CONFLICT(user_id, group_id) DO UPDATE SET
        role = 'admin',
        created_at = COALESCE(group_access.created_at, excluded.created_at)
    WHERE group_access.role != 'owner';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_group_admins_view_delete
INSTEAD OF DELETE ON group_admins
BEGIN
    DELETE FROM group_access
    WHERE user_id = OLD.user_id
      AND group_id = OLD.group_id
      AND role = 'admin';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_group_readers_view_insert
INSTEAD OF INSERT ON group_readers
BEGIN
    INSERT INTO group_access (id, user_id, group_id, role, created_at)
    VALUES (
        COALESCE(NEW.id, 'gac_' || lower(hex(randomblob(10)))),
        NEW.user_id,
        NEW.group_id,
        'viewer',
        COALESCE(NEW.created_at, CURRENT_TIMESTAMP)
    )
    ON CONFLICT(user_id, group_id) DO UPDATE SET
        role = 'viewer',
        created_at = COALESCE(group_access.created_at, excluded.created_at)
    WHERE group_access.role != 'owner';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_group_readers_view_delete
INSTEAD OF DELETE ON group_readers
BEGIN
    DELETE FROM group_access
    WHERE user_id = OLD.user_id
      AND group_id = OLD.group_id
      AND role = 'viewer';
END;
-- +goose StatementEnd
