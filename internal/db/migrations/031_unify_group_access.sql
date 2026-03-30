-- +goose Up
CREATE TABLE IF NOT EXISTS group_access (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'viewer')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    UNIQUE(user_id, group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_access_user_id ON group_access(user_id);
CREATE INDEX IF NOT EXISTS idx_group_access_group_id ON group_access(group_id);
CREATE INDEX IF NOT EXISTS idx_group_access_role ON group_access(role);
CREATE UNIQUE INDEX IF NOT EXISTS idx_group_access_owner_per_group ON group_access(group_id) WHERE role = 'owner';

WITH legacy_memberships AS (
    SELECT
        g.id AS group_id,
        g.admin_user_id AS user_id,
        'admin' AS role,
        g.created_at AS created_at
    FROM groups g

    UNION ALL

    SELECT
        ga.group_id,
        ga.user_id,
        'admin' AS role,
        ga.created_at
    FROM group_admins ga

    UNION ALL

    SELECT
        gr.group_id,
        gr.user_id,
        'viewer' AS role,
        gr.created_at
    FROM group_readers gr
),
deduped_memberships AS (
    SELECT
        lm.group_id,
        lm.user_id,
        CASE
            WHEN SUM(CASE WHEN lm.role = 'admin' THEN 1 ELSE 0 END) > 0 THEN 'admin'
            ELSE 'viewer'
        END AS base_role,
        MIN(lm.created_at) AS membership_created_at
    FROM legacy_memberships lm
    GROUP BY lm.group_id, lm.user_id
),
ranked_memberships AS (
    SELECT
        dm.group_id,
        dm.user_id,
        dm.base_role,
        dm.membership_created_at,
        ROW_NUMBER() OVER (
            PARTITION BY dm.group_id
            ORDER BY dm.membership_created_at ASC, dm.user_id ASC
        ) AS membership_rank
    FROM deduped_memberships dm
)
INSERT INTO group_access (id, user_id, group_id, role, created_at)
SELECT
    'gac_' || lower(hex(randomblob(10))) AS id,
    rm.user_id,
    rm.group_id,
    CASE
        WHEN rm.membership_rank = 1 THEN 'owner'
        ELSE rm.base_role
    END AS role,
    rm.membership_created_at
FROM ranked_memberships rm;

UPDATE groups
SET admin_user_id = (
    SELECT ga.user_id
    FROM group_access ga
    WHERE ga.group_id = groups.id
      AND ga.role = 'owner'
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1
    FROM group_access ga
    WHERE ga.group_id = groups.id
      AND ga.role = 'owner'
);

DROP INDEX IF EXISTS idx_group_admins_group_id;
DROP INDEX IF EXISTS idx_group_admins_user_id;
DROP TABLE IF EXISTS group_admins;

DROP INDEX IF EXISTS idx_group_readers_group_id;
DROP INDEX IF EXISTS idx_group_readers_user_id;
DROP TABLE IF EXISTS group_readers;

DROP TRIGGER IF EXISTS trg_groups_owner_access_insert;
DROP TRIGGER IF EXISTS trg_groups_owner_access_update;
DROP TRIGGER IF EXISTS trg_group_admins_view_insert;
DROP TRIGGER IF EXISTS trg_group_admins_view_delete;
DROP TRIGGER IF EXISTS trg_group_readers_view_insert;
DROP TRIGGER IF EXISTS trg_group_readers_view_delete;
DROP VIEW IF EXISTS group_admins;
DROP VIEW IF EXISTS group_readers;

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

-- +goose StatementBegin
CREATE TRIGGER trg_groups_owner_access_insert
AFTER INSERT ON groups
BEGIN
    INSERT INTO group_access (id, user_id, group_id, role)
    VALUES ('gac_' || lower(hex(randomblob(10))), NEW.admin_user_id, NEW.id, 'owner')
    ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'owner';
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_groups_owner_access_update
AFTER UPDATE OF admin_user_id ON groups
WHEN OLD.admin_user_id != NEW.admin_user_id
BEGIN
    UPDATE group_access
    SET role = 'admin'
    WHERE group_id = NEW.id
      AND user_id = OLD.admin_user_id
      AND role = 'owner';

    INSERT INTO group_access (id, user_id, group_id, role)
    VALUES ('gac_' || lower(hex(randomblob(10))), NEW.admin_user_id, NEW.id, 'owner')
    ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'owner';
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_groups_owner_access_update;
DROP TRIGGER IF EXISTS trg_groups_owner_access_insert;
DROP TRIGGER IF EXISTS trg_group_readers_view_delete;
DROP TRIGGER IF EXISTS trg_group_readers_view_insert;
DROP TRIGGER IF EXISTS trg_group_admins_view_delete;
DROP TRIGGER IF EXISTS trg_group_admins_view_insert;
DROP VIEW IF EXISTS group_readers;
DROP VIEW IF EXISTS group_admins;

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

INSERT INTO group_admins (id, user_id, group_id, created_at)
SELECT
    'gad_' || lower(hex(randomblob(10))) AS id,
    ga.user_id,
    ga.group_id,
    ga.created_at
FROM group_access ga
WHERE ga.role IN ('owner', 'admin');

INSERT INTO group_readers (id, user_id, group_id, created_at)
SELECT
    'grd_' || lower(hex(randomblob(10))) AS id,
    ga.user_id,
    ga.group_id,
    ga.created_at
FROM group_access ga
WHERE ga.role = 'viewer';

UPDATE groups
SET admin_user_id = (
    SELECT ga.user_id
    FROM group_access ga
    WHERE ga.group_id = groups.id
      AND ga.role = 'owner'
    LIMIT 1
)
WHERE EXISTS (
    SELECT 1
    FROM group_access ga
    WHERE ga.group_id = groups.id
      AND ga.role = 'owner'
);

DROP INDEX IF EXISTS idx_group_access_owner_per_group;
DROP INDEX IF EXISTS idx_group_access_role;
DROP INDEX IF EXISTS idx_group_access_group_id;
DROP INDEX IF EXISTS idx_group_access_user_id;
DROP TABLE IF EXISTS group_access;
