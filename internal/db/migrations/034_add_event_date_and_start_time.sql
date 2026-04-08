-- +goose Up
ALTER TABLE events ADD COLUMN date TEXT NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN start_time TEXT NOT NULL DEFAULT '';

UPDATE events
SET date = CASE
        WHEN length(time) >= 10 THEN substr(time, 1, 10)
        ELSE ''
    END,
    start_time = CASE
        WHEN length(time) >= 16 THEN substr(time, 12, 5)
        ELSE '00:00'
    END
WHERE date = '' OR start_time = '';

CREATE INDEX IF NOT EXISTS idx_events_group_date_start_time ON events(group_id, date, start_time);

-- +goose Down
DROP INDEX IF EXISTS idx_events_group_date_start_time;
