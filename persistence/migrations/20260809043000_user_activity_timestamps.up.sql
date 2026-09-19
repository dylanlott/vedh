ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS last_active_at TIMESTAMP;

UPDATE users
SET last_login_at = COALESCE(last_login_at, timestamp),
    last_active_at = COALESCE(last_active_at, last_login_at, timestamp);
