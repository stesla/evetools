-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Default station to Jita IV - Moon 4 - Caldari Navy Assembly Plant
ALTER TABLE users ADD COLUMN stationID INTEGER NOT NULL DEFAULT 60003760;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
-- All this because sqlite doesn't have DROP COLUMN
CREATE TABLE temp AS SELECT id, activeCharacterID FROM users;
DROP TABLE users;
ALTER TABLE temp RENAME TO users;
