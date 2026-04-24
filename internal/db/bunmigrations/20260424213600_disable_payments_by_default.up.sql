INSERT INTO app_flags (key, bool_value, updated_at)
VALUES ('enable_payments', 0, CURRENT_TIMESTAMP)
ON CONFLICT(key) DO UPDATE SET
  bool_value = 0,
  updated_at = CURRENT_TIMESTAMP;
