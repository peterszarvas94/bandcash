-- +goose Up
CREATE TABLE IF NOT EXISTS participants (
    entry_id INTEGER NOT NULL,
    payee_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (entry_id, payee_id),
    FOREIGN KEY (entry_id) REFERENCES entries(id) ON DELETE CASCADE,
    FOREIGN KEY (payee_id) REFERENCES payees(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS participants_entry_id_idx ON participants(entry_id);
CREATE INDEX IF NOT EXISTS participants_payee_id_idx ON participants(payee_id);

-- +goose Down
DROP INDEX IF EXISTS participants_payee_id_idx;
DROP INDEX IF EXISTS participants_entry_id_idx;
DROP TABLE IF EXISTS participants;
