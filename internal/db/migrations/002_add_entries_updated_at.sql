-- +goose Up
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
SELECT id, title, time, description, amount, created_at, created_at FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;

-- +goose Down
CREATE TABLE entries_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    time TEXT NOT NULL,
    description TEXT NOT NULL,
    amount REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO entries_new (id, title, time, description, amount, created_at)
SELECT id, title, time, description, amount, created_at FROM entries;

DROP TABLE entries;
ALTER TABLE entries_new RENAME TO entries;
