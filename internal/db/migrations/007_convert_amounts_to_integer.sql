-- +goose Up
CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    time TEXT NOT NULL,
    description TEXT NOT NULL,
    amount INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO entries_new (id, title, time, description, amount, created_at, updated_at)
SELECT id, title, time, description, CAST(amount AS INTEGER), created_at, updated_at FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS entries_updated_at
AFTER UPDATE ON entries
FOR EACH ROW
BEGIN
    UPDATE entries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

CREATE TABLE participants_new (
    entry_id INTEGER NOT NULL,
    payee_id INTEGER NOT NULL,
    amount INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, payee_id),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
    FOREIGN KEY (payee_id) REFERENCES payees(id) ON DELETE CASCADE
);

INSERT INTO participants_new (entry_id, payee_id, amount, created_at, updated_at)
SELECT entry_id, payee_id, CAST(amount AS INTEGER), created_at, updated_at FROM participants;

DROP TABLE participants;
ALTER TABLE participants_new RENAME TO participants;

CREATE INDEX IF NOT EXISTS participants_entry_id_idx ON participants(entry_id);
CREATE INDEX IF NOT EXISTS participants_payee_id_idx ON participants(payee_id);

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
CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    time TEXT NOT NULL,
    description TEXT NOT NULL,
    amount REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO entries_new (id, title, time, description, amount, created_at, updated_at)
SELECT id, title, time, description, CAST(amount AS REAL), created_at, updated_at FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS entries_updated_at
AFTER UPDATE ON entries
FOR EACH ROW
BEGIN
    UPDATE entries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

CREATE TABLE participants_new (
    entry_id INTEGER NOT NULL,
    payee_id INTEGER NOT NULL,
    amount REAL NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, payee_id),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
    FOREIGN KEY (payee_id) REFERENCES payees(id) ON DELETE CASCADE
);

INSERT INTO participants_new (entry_id, payee_id, amount, created_at, updated_at)
SELECT entry_id, payee_id, CAST(amount AS REAL), created_at, updated_at FROM participants;

DROP TABLE participants;
ALTER TABLE participants_new RENAME TO participants;

CREATE INDEX IF NOT EXISTS participants_entry_id_idx ON participants(entry_id);
CREATE INDEX IF NOT EXISTS participants_payee_id_idx ON participants(payee_id);

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
