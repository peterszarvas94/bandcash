-- +goose Up
ALTER TABLE participants ADD COLUMN amount REAL NOT NULL DEFAULT 0;

-- +goose Down
CREATE TABLE participants_new (
    entry_id INTEGER NOT NULL,
    payee_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, payee_id),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
    FOREIGN KEY (payee_id) REFERENCES payees(id) ON DELETE CASCADE
);

INSERT INTO participants_new (entry_id, payee_id, created_at, updated_at)
SELECT entry_id, payee_id, created_at, updated_at FROM participants;

DROP TABLE participants;
ALTER TABLE participants_new RENAME TO participants;
