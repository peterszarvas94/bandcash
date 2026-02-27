-- +goose Up
CREATE TABLE IF NOT EXISTS expenses (
    id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    amount INTEGER NOT NULL,
    date TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_expenses_group_id ON expenses(group_id);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_expenses_updated_at
AFTER UPDATE ON expenses
FOR EACH ROW
BEGIN
    UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_expenses_updated_at;
DROP TABLE IF EXISTS expenses;
