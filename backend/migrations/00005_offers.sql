-- +goose Up
CREATE TABLE offers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    goal TEXT NOT NULL DEFAULT '',
    default_duration_minutes INTEGER NOT NULL,
    format TEXT NOT NULL,
    price_rubles INTEGER,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMPTZ,
    CONSTRAINT offers_title_not_empty CHECK (BTRIM(title) <> ''),
    CONSTRAINT offers_duration_positive CHECK (default_duration_minutes BETWEEN 15 AND 480),
    CONSTRAINT offers_format_allowed CHECK (format IN ('online', 'offline', 'both')),
    CONSTRAINT offers_price_non_negative CHECK (price_rubles IS NULL OR price_rubles >= 0)
);

CREATE INDEX offers_active_idx ON offers (title) WHERE archived_at IS NULL;

CREATE TABLE teacher_offers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    offer_id BIGINT NOT NULL REFERENCES offers(id) ON DELETE RESTRICT,
    teacher_profile_id BIGINT NOT NULL REFERENCES teacher_profiles(id) ON DELETE RESTRICT,
    duration_minutes INTEGER,
    price_rubles INTEGER,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMPTZ,
    CONSTRAINT teacher_offers_duration_positive CHECK (duration_minutes IS NULL OR duration_minutes BETWEEN 15 AND 480),
    CONSTRAINT teacher_offers_price_non_negative CHECK (price_rubles IS NULL OR price_rubles >= 0)
);

CREATE UNIQUE INDEX teacher_offers_active_pair_unique
    ON teacher_offers (offer_id, teacher_profile_id)
    WHERE archived_at IS NULL;
CREATE INDEX teacher_offers_active_teacher_idx
    ON teacher_offers (teacher_profile_id, offer_id)
    WHERE archived_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS teacher_offers;
DROP TABLE IF EXISTS offers;
