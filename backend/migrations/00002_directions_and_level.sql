-- +goose Up
CREATE TABLE directions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL CHECK (LENGTH(BTRIM(name)) BETWEEN 1 AND 120)
);
CREATE UNIQUE INDEX directions_name_unique ON directions (LOWER(BTRIM(name)));
INSERT INTO directions(name)
SELECT DISTINCT ON (LOWER(BTRIM(title))) BTRIM(title)
FROM slots ORDER BY LOWER(BTRIM(title)), id;
ALTER TABLE slots ADD COLUMN direction_id BIGINT REFERENCES directions(id) ON DELETE RESTRICT;
UPDATE slots SET direction_id = directions.id FROM directions
WHERE LOWER(BTRIM(slots.title)) = LOWER(BTRIM(directions.name));
ALTER TABLE slots ALTER COLUMN direction_id SET NOT NULL;
ALTER TABLE slots DROP COLUMN title;
ALTER TABLE slots ADD COLUMN level TEXT NOT NULL DEFAULT '' CHECK (LENGTH(level) <= 80);
CREATE INDEX slots_direction_idx ON slots(direction_id);

-- +goose Down
ALTER TABLE slots ADD COLUMN title TEXT;
UPDATE slots SET title = directions.name FROM directions WHERE slots.direction_id = directions.id;
ALTER TABLE slots ALTER COLUMN title SET NOT NULL;
ALTER TABLE slots ADD CHECK (LENGTH(BTRIM(title)) BETWEEN 1 AND 120);
ALTER TABLE slots DROP COLUMN direction_id;
ALTER TABLE slots DROP COLUMN level;
DROP TABLE directions;
