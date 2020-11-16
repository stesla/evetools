package model

import (
	"database/sql"
	"errors"
	"golang.org/x/oauth2"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stesla/evetools/esi"
)

type DB interface {
	GetFavoriteTypes(userID int) ([]int, error)
	GetCharacterByOwnerHash(string) (*Character, error)
	GetCharactersForUser(int) (map[int]*Character, error)
	GetUserForVerifiedToken(token *oauth2.Token, verify esi.VerifyOK) (*User, error)
	IsFavorite(userID, typeID int) (bool, error)
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

func (m *databaseModel) GetCharacterByOwnerHash(hash string) (c *Character, err error) {
	const query = `SELECT id, characterID, characterName, userID FROM characters WHERE characterOwnerHash = ?`
	c = &Character{CharacterOwnerHash: hash}
	err = m.db.QueryRow(query, hash).Scan(&c.ID, &c.CharacterID, &c.CharacterName, &c.UserID)
	return
}

func (m *databaseModel) GetCharactersForUser(userID int) (map[int]*Character, error) {
	const query = `SELECT characterID, characterName FROM characters WHERE userID = ?`

	rows, err := m.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	result := map[int]*Character{}
	for rows.Next() {
		c := &Character{UserID: userID}
		if err = rows.Scan(&c.CharacterID, &c.CharacterName); err != nil {
			return nil, err
		}
		result[c.ID] = c
	}
	return result, rows.Err()
}

func (m *databaseModel) GetUserForVerifiedToken(token *oauth2.Token, verify esi.VerifyOK) (u *User, err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	const selectUser = `SELECT u.id, u.activeCharacterHash, c.characterID, u.stationID
						FROM users u JOIN characters c ON u.activeCharacterHash = c.characterOwnerHash
						WHERE c.characterID = ? AND c.characterOwnerHash = ?`
	const createUser = `INSERT INTO users (activeCharacterHash) VALUES (?)`
	const createCharacter = `INSERT INTO characters (characterID, characterName, characterOwnerHash, userID)
							 VALUES (?, ?, ?, ?)`

	u = &User{}
	err = tx.QueryRow(selectUser, verify.CharacterID, verify.CharacterOwnerHash).
		Scan(&u.ID, &u.ActiveCharacterHash, &u.ActiveCharacterID, &u.StationID)
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

		_, err = tx.Exec(createCharacter, verify.CharacterID, verify.CharacterName, verify.CharacterOwnerHash, userID)
		if err != nil {
			return
		}

		u.ID = int(userID)
		u.ActiveCharacterHash = verify.CharacterOwnerHash
		u.ActiveCharacterID = verify.CharacterID
		u.StationID = 60003760 // default value in the database
	}
	return
}

func (m *databaseModel) SaveUserStation(userID, stationID int) error {
	const query = `UPDATE users SET stationID = ? WHERE id = ?`
	_, err := m.db.Exec(query, stationID, userID)
	return err
}
