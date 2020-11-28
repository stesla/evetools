-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE prices (
  stationID  INTEGER  NOT NULL,
  typeID     INTEGER  NOT NULL,
  buy        REAL     NOT NULL,
  sell       REAL     NOT NULL,

  PRIMARY KEY (stationID, typeID) ON CONFLICT REPLACE
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE prices;
