-- Supports the case-insensitive email lookup in GetUserByEmail.
-- Not UNIQUE: rows created before emails were normalised may differ only by case.
CREATE INDEX IF NOT EXISTS users_email_lower_idx ON users (LOWER(email));
