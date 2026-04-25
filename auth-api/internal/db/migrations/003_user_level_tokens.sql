-- Token buckets are now per-user (shared across all API keys of a user).
-- The redis key changed from rl:<sha256> to rl:user:<user_id>.
-- Existing Redis keys will expire naturally; new buckets are created on first use.

-- Update the column default so newly created keys record the monthly interval.
ALTER TABLE api_keys
    ALTER COLUMN tokens_reset_interval_secs SET DEFAULT 2592000;

-- Migrate existing rows that still have the old hourly default (3600 s).
UPDATE api_keys
    SET tokens_reset_interval_secs = 2592000
    WHERE tokens_reset_interval_secs = 3600;
