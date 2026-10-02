-- +goose Up
ALTER TABLE lessons ADD COLUMN description TEXT NOT NULL DEFAULT '';

CREATE TABLE applications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE RESTRICT,
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'new',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMPTZ,
    CONSTRAINT applications_full_name_not_empty CHECK (BTRIM(full_name) <> ''),
    CONSTRAINT applications_phone_not_empty CHECK (BTRIM(phone) <> ''),
    CONSTRAINT applications_status_allowed CHECK (status IN ('new', 'accepted', 'rejected', 'completed'))
);

CREATE INDEX applications_active_lesson_idx
    ON applications (lesson_id, status, created_at)
    WHERE archived_at IS NULL;
CREATE INDEX applications_active_created_idx
    ON applications (created_at DESC)
    WHERE archived_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS applications;
ALTER TABLE lessons DROP COLUMN IF EXISTS description;
