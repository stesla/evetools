-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE types (
  typeID INTEGER PRIMARY KEY ASC,
  favorite BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE types;
