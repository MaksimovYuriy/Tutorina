-- +goose Up
ALTER TABLE teacher_profiles ADD COLUMN archived_at TIMESTAMPTZ;
CREATE INDEX teacher_profiles_active_idx ON teacher_profiles (display_name) WHERE archived_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS teacher_profiles_active_idx;
ALTER TABLE teacher_profiles DROP COLUMN archived_at;

