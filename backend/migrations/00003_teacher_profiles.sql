-- +goose Up
ALTER TABLE teacher_profiles
    ALTER COLUMN user_id DROP NOT NULL,
    ADD COLUMN education TEXT NOT NULL DEFAULT '',
    ADD COLUMN experience TEXT NOT NULL DEFAULT '',
    ADD COLUMN approach TEXT NOT NULL DEFAULT '',
    ADD COLUMN photo_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE teacher_profiles
    DROP COLUMN photo_url,
    DROP COLUMN approach,
    DROP COLUMN experience,
    DROP COLUMN education,
    ALTER COLUMN user_id SET NOT NULL;

