ALTER TABLE events ADD COLUMN IF NOT EXISTS date TEXT NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN IF NOT EXISTS event_time TEXT NOT NULL DEFAULT '';

UPDATE events
SET date = substr(time, 1, 10)
WHERE trim(COALESCE(date, '')) = ''
  AND length(time) >= 10;

UPDATE events
SET event_time = substr(time, 12, 5)
WHERE trim(COALESCE(event_time, '')) = ''
  AND length(time) >= 16;
