-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE users (
  id                   INTEGER  PRIMARY KEY AUTOINCREMENT,
  activeCharacterHash  TEXT     NOT NULL,
  -- default: Jita IV - Moon 4 - Caldari Navy Assembly Plant
  stationID            INTEGER  NOT NULL DEFAULT 60003760,

  FOREIGN KEY(activeCharacterHash) REFERENCES characters(characterOwnerHash)
);

CREATE TABLE characters (
  id                  INTEGER  PRIMARY KEY AUTOINCREMENT,
  characterID         INTEGER  NOT NULL,
  characterName       TEXT     NOT NULL,
  characterOwnerHash  TEXT     NOT NULL,
  userID              INTEGER  NOT NULL,

  FOREIGN KEY(userID) REFERENCES users(id),
  UNIQUE(characterID, characterOwnerHash) ON CONFLICT REPLACE
);

CREATE TABLE tokens (
  id            INTEGER  PRIMARY KEY AUTOINCREMENT,
  characterID   INTEGER  NOT NULL,
  refreshToken  TEXT     NOT NULL,
  scopes        TEXT     NOT NULL,

  UNIQUE(characterID) ON CONFLICT REPLACE,
  FOREIGN KEY(characterID) REFERENCES characters(id)
);

CREATE TABLE favorites (
  userID    INTEGER  NOT NULL,
  typeID    INTEGER  NOT NULL,

  FOREIGN KEY (userID) REFERENCES user(id),
  PRIMARY KEY(userID, typeID) ON CONFLICT IGNORE
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE favorites;
DROP TABLE scopes;
DROP TABLE tokens;
DROP TABLE characters;
DROP TABLE users;
