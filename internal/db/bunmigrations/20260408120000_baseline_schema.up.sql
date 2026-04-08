PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    preferred_lang TEXT NOT NULL DEFAULT 'hu'
);

CREATE TABLE IF NOT EXISTS groups (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    admin_user_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS magic_links (
    id TEXT PRIMARY KEY,
    token TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL,
    action TEXT NOT NULL,
    group_id TEXT,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    invite_role TEXT NOT NULL DEFAULT 'viewer' CHECK (invite_role IN ('viewer', 'admin')),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_magic_links_token ON magic_links(token);
CREATE INDEX IF NOT EXISTS idx_magic_links_email ON magic_links(email);

CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    title TEXT NOT NULL,
    time TEXT NOT NULL,
    description TEXT NOT NULL,
    amount INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    paid INTEGER NOT NULL DEFAULT 0,
    paid_at TEXT,
    place TEXT NOT NULL DEFAULT '',
    date TEXT NOT NULL DEFAULT '',
    event_time TEXT NOT NULL DEFAULT '',
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_events_group_id ON events(group_id);

CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_members_group_id ON members(group_id);

CREATE TABLE IF NOT EXISTS participants (
    group_id TEXT NOT NULL,
    event_id TEXT NOT NULL,
    member_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    expense INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    paid INTEGER NOT NULL DEFAULT 0,
    paid_at TEXT,
    note TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (event_id, member_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_participants_group_id ON participants(group_id);
CREATE INDEX IF NOT EXISTS idx_participants_event_id ON participants(event_id);
CREATE INDEX IF NOT EXISTS idx_participants_member_id ON participants(member_id);

CREATE TABLE IF NOT EXISTS app_flags (
  key TEXT PRIMARY KEY,
  bool_value INTEGER NOT NULL CHECK (bool_value IN (0, 1)),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS banned_users (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL UNIQUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_banned_users_user_id ON banned_users(user_id);

CREATE TABLE IF NOT EXISTS expenses (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    amount INTEGER NOT NULL,
    date TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    paid INTEGER NOT NULL DEFAULT 0,
    paid_at TEXT,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_expenses_group_id ON expenses(group_id);

CREATE TABLE IF NOT EXISTS user_detail_card_states (
  user_id TEXT NOT NULL,
  state_key TEXT NOT NULL,
  is_open INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, state_key),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_detail_card_states_user_id ON user_detail_card_states(user_id);

CREATE TABLE IF NOT EXISTS user_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);

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

CREATE TRIGGER IF NOT EXISTS trg_events_updated_at
AFTER UPDATE ON events
FOR EACH ROW
BEGIN
    UPDATE events SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS trg_members_updated_at
AFTER UPDATE ON members
FOR EACH ROW
BEGIN
    UPDATE members SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS trg_participants_updated_at
AFTER UPDATE ON participants
FOR EACH ROW
BEGIN
    UPDATE participants
    SET updated_at = CURRENT_TIMESTAMP
    WHERE event_id = NEW.event_id AND member_id = NEW.member_id;
END;

CREATE TRIGGER IF NOT EXISTS trg_expenses_updated_at
AFTER UPDATE ON expenses
FOR EACH ROW
BEGIN
    UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS trg_groups_owner_access_insert
AFTER INSERT ON groups
BEGIN
    INSERT INTO group_access (id, user_id, group_id, role)
    VALUES ('gac_' || lower(hex(randomblob(10))), NEW.admin_user_id, NEW.id, 'owner')
    ON CONFLICT(user_id, group_id) DO UPDATE SET role = 'owner';
END;

CREATE TRIGGER IF NOT EXISTS trg_groups_owner_access_update
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

CREATE VIEW IF NOT EXISTS group_outgoing_payments AS
SELECT
  p.group_id AS group_id,
  'participant' AS payment_kind,
  CAST(p.event_id || ':' || p.member_id AS TEXT) AS payment_id,
  CAST(p.event_id AS TEXT) AS event_id,
  CAST(p.member_id AS TEXT) AS member_id,
  CAST(m.name AS TEXT) AS member_name,
  CAST(e.title AS TEXT) AS event_title,
  e.title AS title,
  CAST((p.amount + p.expense) AS INTEGER) AS amount,
  p.paid AS paid,
  p.paid_at AS paid_at,
  p.updated_at AS updated_at,
  e.time AS sort_date
FROM participants p
JOIN members m ON m.id = p.member_id AND m.group_id = p.group_id
JOIN events e ON e.id = p.event_id AND e.group_id = p.group_id
UNION ALL
SELECT
  ex.group_id AS group_id,
  'expense' AS payment_kind,
  CAST(ex.id AS TEXT) AS payment_id,
  '' AS event_id,
  '' AS member_id,
  '' AS member_name,
  '' AS event_title,
  ex.title AS title,
  CAST(ex.amount AS INTEGER) AS amount,
  ex.paid AS paid,
  ex.paid_at AS paid_at,
  ex.updated_at AS updated_at,
  ex.date AS sort_date
FROM expenses ex;
