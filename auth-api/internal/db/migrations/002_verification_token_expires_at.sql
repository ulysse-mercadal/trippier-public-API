ALTER TABLE users
    ADD COLUMN IF NOT EXISTS verification_token_expires_at TIMESTAMPTZ;
