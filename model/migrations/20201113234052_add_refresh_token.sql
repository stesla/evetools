-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- Default station to Jita IV - Moon 4 - Caldari Navy Assembly Plant
ALTER TABLE characters ADD COLUMN refreshToken TEXT;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
-- All this because sqlite doesn't have DROP COLUMN
CREATE TABLE temp AS SELECT id, characterID, characterName, owner, userID FROM characters;
DROP TABLE characters;
ALTER TABLE temp RENAME TO characters;
