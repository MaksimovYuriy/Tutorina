-- +goose Up
CREATE TABLE users (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 username TEXT NOT NULL UNIQUE CHECK (username ~ '^[a-z0-9_-]{3,64}$'),
 password_hash TEXT NOT NULL,
 is_active BOOLEAN NOT NULL DEFAULT TRUE,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE roles (id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY, code TEXT NOT NULL UNIQUE CHECK (code = 'admin'));
INSERT INTO roles (code) VALUES ('admin');
CREATE TABLE user_roles (
 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 role_id SMALLINT NOT NULL REFERENCES roles(id), PRIMARY KEY(user_id, role_id)
);
CREATE TABLE sessions (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token_hash BYTEA NOT NULL UNIQUE CHECK (OCTET_LENGTH(token_hash) = 32),
 expires_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 revoked_at TIMESTAMPTZ,
 CHECK (expires_at > created_at)
);
CREATE INDEX sessions_active_user_idx ON sessions(user_id, expires_at) WHERE revoked_at IS NULL;
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
DROP TABLE user_roles;
DROP TABLE roles;
DROP TABLE users;
