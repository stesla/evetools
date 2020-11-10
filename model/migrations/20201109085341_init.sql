-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE users (
  id                 INTEGER  PRIMARY KEY AUTOINCREMENT,
  activeCharacterID  INTEGER  NOT NULL
);

CREATE TABLE characters (
  id             INTEGER  PRIMARY KEY AUTOINCREMENT,
  characterID    INTEGER  NOT NULL,
  characterName  TEXT     NOT NULL,
  owner          TEXT     NOT NULL,
  userID         INTEGER  NOT NULL,

  FOREIGN KEY(userID) REFERENCES users(id),
  UNIQUE(characterID, owner)
);

CREATE TABLE favorites (
  userID    INTEGER  NOT NULL,
  typeID    INTEGER  NOT NULL,

  PRIMARY KEY(userID, typeID) ON CONFLICT IGNORE
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE favorites;
DROP TABLE characters;
DROP TABLE users;
