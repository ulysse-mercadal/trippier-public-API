-- Quota is per-user (the Redis bucket has been rl:user:<id> since 003), so the
-- columns on api_keys never made sense and were never read at runtime. Move the
-- quota to users where the bucket already lives.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tokens_limit               INTEGER NOT NULL DEFAULT 1000,
    ADD COLUMN IF NOT EXISTS tokens_reset_interval_secs INTEGER NOT NULL DEFAULT 2592000;

ALTER TABLE api_keys
    DROP COLUMN IF EXISTS tokens_limit,
    DROP COLUMN IF EXISTS tokens_reset_interval_secs;
