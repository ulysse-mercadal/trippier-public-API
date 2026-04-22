CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email              VARCHAR(255) UNIQUE NOT NULL,
    password_hash      VARCHAR(255) NOT NULL,
    verified           BOOLEAN      NOT NULL DEFAULT false,
    verification_token VARCHAR(64),
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id                          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                        VARCHAR(255) NOT NULL,
    key_hash_bcrypt             VARCHAR(255) NOT NULL,
    key_hash_sha256             CHAR(64)     NOT NULL UNIQUE,
    key_prefix                  VARCHAR(16)  NOT NULL,
    tokens_limit                INTEGER      NOT NULL DEFAULT 1000,
    tokens_reset_interval_secs  INTEGER      NOT NULL DEFAULT 3600,
    revoked                     BOOLEAN      NOT NULL DEFAULT false,
    created_at                  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_used_at                TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id    ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_sha256     ON api_keys(key_hash_sha256);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix     ON api_keys(key_prefix);
