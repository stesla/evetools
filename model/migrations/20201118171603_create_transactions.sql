-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE wallet_transactions (
  txnID         INTEGER PRIMARY KEY,
  clientID      INTEGER NOT NULL,
  clientName    TEXT NOT NULL,
  date          TEXT NOT NULL,
  isBuy         BOOLEAN NOT NULL,
  isPersonal    BOOLEAN NOT NULL,
  journalRefID  INTEGER NOT NULL,
  locationID    INTEGER NOT NULL,
  quantity      INTEGER NOT NULL,
  typeID        INTEGER NOT NULL,
  unitPrice     REAL NOT NULL
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE wallet_transactions;
