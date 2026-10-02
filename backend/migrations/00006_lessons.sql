-- +goose Up
CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE teacher_offers
    ADD CONSTRAINT teacher_offers_id_teacher_unique UNIQUE (id, teacher_profile_id);

CREATE TABLE lessons (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    teacher_offer_id BIGINT NOT NULL,
    teacher_profile_id BIGINT NOT NULL,
    offer_title TEXT NOT NULL,
    price_rubles INTEGER,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    delivery_format TEXT NOT NULL,
    lesson_type TEXT NOT NULL,
    capacity INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'planned',
    enrollment_open BOOLEAN NOT NULL DEFAULT TRUE,
    group_goal TEXT NOT NULL DEFAULT '',
    group_level TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    archived_at TIMESTAMPTZ,
    CONSTRAINT lessons_teacher_offer_fk
        FOREIGN KEY (teacher_offer_id, teacher_profile_id)
        REFERENCES teacher_offers(id, teacher_profile_id) ON DELETE RESTRICT,
    CONSTRAINT lessons_time_order CHECK (ends_at > starts_at),
    CONSTRAINT lessons_delivery_format_allowed CHECK (delivery_format IN ('online', 'offline')),
    CONSTRAINT lessons_type_allowed CHECK (lesson_type IN ('individual', 'group')),
    CONSTRAINT lessons_capacity_by_type CHECK (
        (lesson_type = 'individual' AND capacity = 1) OR
        (lesson_type = 'group' AND capacity >= 2)
    ),
    CONSTRAINT lessons_status_allowed CHECK (status IN ('planned', 'completed', 'cancelled')),
    CONSTRAINT lessons_price_non_negative CHECK (price_rubles IS NULL OR price_rubles >= 0),
    CONSTRAINT lessons_teacher_no_overlap EXCLUDE USING gist (
        teacher_profile_id WITH =,
        tstzrange(starts_at, ends_at, '[)') WITH &&
    ) WHERE (archived_at IS NULL AND status <> 'cancelled')
);

CREATE INDEX lessons_active_time_idx ON lessons (starts_at, ends_at) WHERE archived_at IS NULL;
CREATE INDEX lessons_active_teacher_time_idx ON lessons (teacher_profile_id, starts_at) WHERE archived_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS lessons;
ALTER TABLE teacher_offers DROP CONSTRAINT IF EXISTS teacher_offers_id_teacher_unique;
