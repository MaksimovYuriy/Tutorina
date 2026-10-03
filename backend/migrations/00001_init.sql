-- +goose Up
CREATE TABLE access_keys (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    key_hash BYTEA NOT NULL UNIQUE CHECK (OCTET_LENGTH(key_hash) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE sessions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key_hash BYTEA NOT NULL REFERENCES access_keys(key_hash) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE CHECK (OCTET_LENGTH(token_hash) = 32),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ,
    CHECK (expires_at > created_at)
);
CREATE INDEX sessions_active_idx ON sessions(expires_at) WHERE revoked_at IS NULL;
CREATE TABLE slots (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 title TEXT NOT NULL CHECK (LENGTH(BTRIM(title)) BETWEEN 1 AND 120),
 starts_at TIMESTAMPTZ NOT NULL,
 ends_at TIMESTAMPTZ NOT NULL CHECK (ends_at > starts_at),
 format TEXT NOT NULL CHECK (format IN ('online','offline')),
 kind TEXT NOT NULL CHECK (kind IN ('individual','group')),
 capacity INTEGER NOT NULL CHECK (capacity BETWEEN 1 AND 1000),
 occupied INTEGER NOT NULL CHECK (occupied BETWEEN 0 AND capacity),
 status TEXT NOT NULL CHECK (status IN ('planned','completed','cancelled')),
 published BOOLEAN NOT NULL DEFAULT FALSE,
 CHECK (kind <> 'individual' OR capacity = 1)
);
CREATE INDEX slots_public_time_idx ON slots(starts_at) WHERE published AND status = 'planned';
-- +goose Down
DROP TABLE slots;
DROP TABLE sessions;
DROP TABLE access_keys;
