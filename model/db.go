package model

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	FavoriteTypes(userID int) ([]int, error)
	FindOrCreateUserForCharacter(characterID int, characterName, owner string) (*User, error)
	GetCharacter(int) (*Character, error)
	IsFavorite(userID, typeID int) (bool, error)
	SaveUserStation(userID, stationID int) error
	SetFavorite(userID, typeID int, val bool) error
}

var ErrNotFound = errors.New("Not Found")

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

func (m *databaseModel) FavoriteTypes(userID int) ([]int, error) {
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
		_, err = m.db.Exec("DELETE FROM favorites WHERE typeID = ? AND userID = ?", val, typeID, userID)
	}
	return
}

type Character struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Owner  string `json:"-"`
	UserID string `json:"-"`
}

type User struct {
	ID                int `json:"id"`
	ActiveCharacterID int `json:"activeCharacterID"`
	StationID         int `json:"stationID"`
}

func (m *databaseModel) FindOrCreateUserForCharacter(characterID int, characterName, owner string) (*User, error) {
	const getUserID = `SELECT c.userID, u.stationID
					   FROM characters AS c JOIN users AS u ON c.userID = u.ID
					   WHERE characterID = ? AND owner = ?`
	const createCharacter = `INSERT INTO characters 
							 (characterID, characterName, owner, userID)
							 VALUES (?, ?, ?, ?)`
	const createUser = `INSERT INTO users (activeCharacterID) VALUES (?)`

	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}

	var userID int
	var stationID int = 60003760 // Jita IV - Moon 4 - Caldari Navy Assembly Plant
	if err = tx.QueryRow(getUserID, characterID, owner).Scan(&userID, &stationID); err == sql.ErrNoRows {
		var r sql.Result
		r, err = tx.Exec(createUser, characterID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		var id int64
		id, err = r.LastInsertId()
		userID = int(id)

		_, err = tx.Exec(createCharacter, characterID, characterName, owner, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}
	return &User{
		ID:                userID,
		ActiveCharacterID: characterID,
		StationID:         stationID,
	}, tx.Commit()
}

func (m *databaseModel) GetCharacter(characterID int) (*Character, error) {
	const query = `SELECT characterName, owner, userID FROM characters WHERE characterID = ?`

	c := &Character{ID: characterID}
	err := m.db.QueryRow(query, characterID).Scan(&c.Name, &c.Owner, &c.UserID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return c, nil
}

func (m *databaseModel) SaveUserStation(userID, stationID int) error {
	const query = `UPDATE users SET stationID = ? WHERE id = ?`
	_, err := m.db.Exec(query, stationID, userID)
	return err
}
