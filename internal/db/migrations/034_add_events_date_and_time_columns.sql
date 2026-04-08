-- +goose Up
ALTER TABLE events ADD COLUMN date TEXT NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN event_time TEXT NOT NULL DEFAULT '';

UPDATE events
SET
  date = CASE
    WHEN trim(time) = '' THEN ''
    ELSE COALESCE(NULLIF(strftime('%Y-%m-%d', replace(time, 'T', ' ')), ''), substr(time, 1, 10), '')
  END,
  event_time = CASE
    WHEN trim(time) = '' THEN ''
    ELSE COALESCE(NULLIF(strftime('%H:%M', replace(time, 'T', ' ')), ''), substr(time, 12, 5), '')
  END
WHERE date = '' OR event_time = '';

-- +goose Down
ALTER TABLE events DROP COLUMN event_time;
ALTER TABLE events DROP COLUMN date;
