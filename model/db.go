package model

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stesla/evetools/esi"
)

type DB interface {
	FindOrCreateCharacterForUser(int, esi.VerifyOK) (*Character, error)
	FindOrCreateUserAndCharacter(esi.VerifyOK) (*User, *Character, error)
	GetCharacterByUserAndCharacterID(int, int) (*Character, error)
	GetCharacterByOwnerHash(string) (*Character, error)
	GetCharactersForUser(int) (map[int]*Character, error)
	GetFavoriteTypes(userID int) ([]int, error)
	GetLatestTxnID() (int, error)
	GetTokenForCharacter(characterID int) (*Token, error)
	GetTransactions() ([]*esi.WalletTransaction, error)
	GetUser(userID int) (*User, error)
	IsFavorite(userID, typeID int) (bool, error)
	SaveActiveCharacterHash(int, string) error
	SaveTokenForCharacter(int, string, string) error
	SaveTransaction(*esi.WalletTransaction) error
	SaveUserStation(userID, stationID int) error
	SetFavorite(userID, typeID int, val bool) error
}

var (
	ErrNotFound       = errors.New("Not Found")
	ErrNotImplemented = errors.New("Not Implemented")
)

type databaseModel struct {
	db *sql.DB
}

func Initialize(dbfilename string) (DB, error) {
	db, err := sql.Open("sqlite3", dbfilename)
	if err != nil {
		return nil, err
	}
	return &databaseModel{db: db}, nil
}

func (m *databaseModel) GetFavoriteTypes(userID int) ([]int, error) {
	const query = "SELECT typeID FROM favorites WHERE userID = ?"
	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	result := []int{}
	for rows.Next() {
		var typeID int
		if err := rows.Scan(&typeID); err != nil {
			return nil, err
		}
		result = append(result, typeID)
	}
	return result, rows.Err()
}

func (m *databaseModel) IsFavorite(userID, typeID int) (bool, error) {
	const query = "SELECT typeID FROM favorites WHERE userID = ? and typeID = ?"
	var id int
	err := m.db.QueryRow(query, userID, typeID).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (m *databaseModel) SetFavorite(userID int, typeID int, val bool) (err error) {
	if val {
		_, err = m.db.Exec("INSERT INTO favorites (userID, typeID) VALUES (?, ?)", userID, typeID)
	} else {
		_, err = m.db.Exec("DELETE FROM favorites WHERE userID = ? AND typeID = ?", userID, typeID)
	}
	return
}

type Character struct {
	ID                 int    `json:"-"`
	CharacterID        int    `json:"id"`
	CharacterName      string `json:"name"`
	CharacterOwnerHash string `json:"-"`
	UserID             int    `json:"-"`
}

type User struct {
	ID                  int    `json:"-"`
	ActiveCharacterHash string `json:"-"`

	ActiveCharacterID int `json:"activeCharacterID"`
	StationID         int `json:"stationID"`
}

const createCharacter = `INSERT INTO characters 
	                     (characterID, characterName, characterOwnerHash, userID)
						 VALUES (?, ?, ?, ?)`

func (m *databaseModel) FindOrCreateCharacterForUser(userID int, verify esi.VerifyOK) (c *Character, err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	const selectCharacter = `SELECT id, characterID, characterName, userID FROM characters
							 WHERE characterOwnerHash = ?`

	c = &Character{CharacterOwnerHash: verify.CharacterOwnerHash}
	err = tx.QueryRow(selectCharacter, verify.CharacterOwnerHash).Scan(
		&c.ID, &c.CharacterID, &c.CharacterName, &c.UserID)
	if err == sql.ErrNoRows {
		var r sql.Result
		r, err = tx.Exec(createCharacter,
			verify.CharacterID, verify.CharacterName,
			verify.CharacterOwnerHash, userID)
		if err != nil {
			return
		}
		var cID int64
		cID, err = r.LastInsertId()
		if err != nil {
			return
		}
		c = &Character{
			ID:                 int(cID),
			CharacterID:        verify.CharacterID,
			CharacterName:      verify.CharacterName,
			CharacterOwnerHash: verify.CharacterOwnerHash,
			UserID:             userID,
		}
	} else if err != nil {
		return
	} else if c.UserID != userID {
		err = errors.New("user already associated with another user")
	}
	return
}

func (m *databaseModel) FindOrCreateUserAndCharacter(verify esi.VerifyOK) (user *User, character *Character, err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	const createUser = `INSERT INTO users (activeCharacterHash) VALUES (?)`
	const selectCharacter = `SELECT id, characterID, characterName, userID
							 FROM characters WHERE characterOwnerHash = ?`
	const selectUser = `SELECT u.activeCharacterHash, c.characterID, u.stationA
                        FROM users u JOIN characters c ON u.activeCharacterHash = c.characterOwnerHash
                        WHERE u.id = ?`

	character = &Character{CharacterOwnerHash: verify.CharacterOwnerHash}
	err = tx.QueryRow(selectCharacter, verify.CharacterOwnerHash).Scan(
		&character.ID, &character.CharacterID,
		&character.CharacterName, &character.UserID)
	if err == sql.ErrNoRows {
		var r sql.Result
		r, err = tx.Exec(createUser, verify.CharacterOwnerHash)
		if err != nil {
			return
		}
		var userID int64
		userID, err = r.LastInsertId()
		if err != nil {
			return
		}

		r, err = tx.Exec(createCharacter, verify.CharacterID, verify.CharacterName, verify.CharacterOwnerHash, userID)
		if err != nil {
			return
		}
		var cID int64
		cID, err = r.LastInsertId()
		if err != nil {
			return
		}

		user = &User{
			ID:                  int(userID),
			ActiveCharacterHash: verify.CharacterOwnerHash,
			ActiveCharacterID:   verify.CharacterID,
			// Jita IV - Moon 4 - Caldari Navy Assembly Plant
			StationID: 60003760,
		}

		character = &Character{
			ID:                 int(cID),
			CharacterID:        verify.CharacterID,
			CharacterName:      verify.CharacterName,
			CharacterOwnerHash: verify.CharacterOwnerHash,
			UserID:             int(userID),
		}
	} else {
		user = &User{ID: character.UserID}
		err = tx.QueryRow(selectUser, character.UserID).Scan(
			&user.ActiveCharacterHash, &user.ActiveCharacterID, &user.StationID)
		if err != nil {
			return
		}
	}
	return
}

func (m *databaseModel) GetCharacterByUserAndCharacterID(userID int, characterID int) (c *Character, err error) {
	const query = `SELECT id, characterName, characterOwnerHash FROM characters
				 WHERE userID = ? and characterID = ?`
	c = &Character{CharacterID: characterID, UserID: userID}
	err = m.db.QueryRow(query, userID, characterID).Scan(&c.ID, &c.CharacterName, &c.CharacterOwnerHash)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func (m *databaseModel) GetCharacterByOwnerHash(hash string) (*Character, error) {
	const query = `SELECT id, characterID, characterName, userID FROM characters WHERE characterOwnerHash = ?`
	c := &Character{CharacterOwnerHash: hash}
	err := m.db.QueryRow(query, hash).Scan(&c.ID, &c.CharacterID, &c.CharacterName, &c.UserID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return c, nil
}

func (m *databaseModel) GetCharactersForUser(userID int) (map[int]*Character, error) {
	const query = `SELECT id, characterID, characterName FROM characters WHERE userID = ?`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	result := map[int]*Character{}
	for rows.Next() {
		c := &Character{UserID: userID}
		if err = rows.Scan(&c.ID, &c.CharacterID, &c.CharacterName); err != nil {
			return nil, err
		}
		result[c.CharacterID] = c
	}
	return result, rows.Err()
}

type Token struct {
	ID           int
	CharacterID  int
	RefreshToken string
	Scopes       string
}

func (m *databaseModel) GetLatestTxnID() (int, error) {
	const query = `SELECT MAX(txnID) FROM wallet_transactions`
	var id sql.NullInt64
	err := m.db.QueryRow(query).Scan(&id)
	return int(id.Int64), err
}

func (m *databaseModel) GetTokenForCharacter(characterID int) (token *Token, err error) {
	const selectToken = `SELECT id, refreshToken, scopes FROM tokens WHERE characterID = ?`
	token = &Token{CharacterID: characterID}
	err = m.db.QueryRow(selectToken, characterID).Scan(
		&token.ID, &token.RefreshToken, &token.Scopes)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func (m *databaseModel) GetTransactions() ([]*esi.WalletTransaction, error) {
	const query = `SELECT txnID, clientID, clientName, date, isBuy, isPersonal,
				          journalRefID, locationID, quantity, typeID, unitPrice
				   FROM wallet_transactions ORDER BY txnID DESC`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	result := []*esi.WalletTransaction{}
	for rows.Next() {
		var t esi.WalletTransaction
		err = rows.Scan(
			&t.TransactionID, &t.ClientID, &t.ClientName, &t.Date, &t.IsBuy, &t.IsPersonal,
			&t.JournalRefID, &t.LocationID, &t.Quantity, &t.TypeID, &t.UnitPrice)
		if err != nil {
			return nil, err
		}
		result = append(result, &t)
	}
	return result, rows.Err()
}

func (m *databaseModel) GetUser(userID int) (*User, error) {
	const selectUser = `SELECT u.activeCharacterHash, c.characterID, u.stationA
                        FROM users u JOIN characters c ON u.activeCharacterHash = c.characterOwnerHash
                        WHERE u.id = ?`
	u := &User{ID: userID}
	err := m.db.QueryRow(selectUser, userID).
		Scan(&u.ActiveCharacterHash, &u.ActiveCharacterID, &u.StationID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return u, nil
}

func (m *databaseModel) SaveActiveCharacterHash(userID int, hash string) (err error) {
	const query = `UPDATE users SET activeCharacterHash = ? WHERE id = ?`
	_, err = m.db.Exec(query, hash, userID)
	return
}

func (m *databaseModel) SaveTokenForCharacter(characterID int, scopes, token string) (err error) {
	const createToken = `INSERT INTO tokens (characterID, refreshToken, scopes) VALUES (?, ?, ?)`
	_, err = m.db.Exec(createToken, characterID, token, scopes)
	return
}

func (m *databaseModel) SaveTransaction(t *esi.WalletTransaction) (err error) {
	const query = `INSERT INTO wallet_transactions
				 (txnID, clientID, clientName, date, isBuy, isPersonal,
				 journalRefID, locationID, quantity, typeID, unitPrice)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = m.db.Exec(query,
		t.TransactionID, t.ClientID, t.ClientName, t.Date, t.IsBuy, t.IsPersonal,
		t.JournalRefID, t.LocationID, t.Quantity, t.TypeID, t.UnitPrice)
	return
}

func (m *databaseModel) SaveUserStation(userID, stationID int) error {
	const query = `UPDATE users SET stationA = ? WHERE id = ?`
	_, err := m.db.Exec(query, stationID, userID)
	return err
}
