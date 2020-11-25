-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE temp (
  id                   INTEGER  PRIMARY KEY AUTOINCREMENT,
  activeCharacterHash  TEXT     NOT NULL,
  -- default: Jita IV - Moon 4 - Caldari Navy Assembly Plant
  stationA             INTEGER  NOT NULL DEFAULT 60003760,
  -- default: Amarr VIII (Oris) - Emperor Family Academy
  stationB             INTEGER  NOT NULL DEFAULT 60008494,

  FOREIGN KEY(activeCharacterHash) REFERENCES characters(characterOwnerHash)
);

INSERT INTO temp(id, activeCharacterHash, stationA)
SELECT id, activeCharacterHash, stationID
FROM users;

DROP TABLE users;

ALTER TABLE temp RENAME TO users;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

CREATE TABLE temp (
  id                   INTEGER  PRIMARY KEY AUTOINCREMENT,
  activeCharacterHash  TEXT     NOT NULL,
  -- default: Jita IV - Moon 4 - Caldari Navy Assembly Plant
  stationID            INTEGER  NOT NULL DEFAULT 60003760,

  FOREIGN KEY(activeCharacterHash) REFERENCES characters(characterOwnerHash)
);

INSERT INTO temp(id, activeCharacterHash, stationID)
SELECT id, activeCharacterHash, stationA
FROM users;

DROP TABLE users;

ALTER TABLE temp RENAME to users;
