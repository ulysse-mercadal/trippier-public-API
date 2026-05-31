-- Token buckets are now per-user (shared across all API keys of a user).
-- The redis key changed from rl:<sha256> to rl:user:<user_id>.
-- Existing Redis keys will expire naturally; new buckets are created on first use.
--
-- Migration 004 later drops the api_keys.tokens_reset_interval_secs column
-- entirely. The migration runner re-executes every file on every boot, so this
-- migration must no-op when the column is already gone.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name   = 'api_keys'
          AND column_name  = 'tokens_reset_interval_secs'
    ) THEN
        ALTER TABLE api_keys
            ALTER COLUMN tokens_reset_interval_secs SET DEFAULT 2592000;

        UPDATE api_keys
            SET tokens_reset_interval_secs = 2592000
            WHERE tokens_reset_interval_secs = 3600;
    END IF;
END $$;
