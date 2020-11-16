package model

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	GetFavoriteTypes(userID int) ([]int, error)
	GetCharacter(int) (*Character, error)
	GetCharactersForUser(int) (map[int]*Character, error)
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
	ID           int    `json:"id"`
	Name         string `json:"name"`
	OwnerHash    string `json:"-"`
	UserID       int    `json:"-"`
	RefreshToken string `json:"-"`
}

type User struct {
	ID                int `json:"id"`
	ActiveCharacterID int `json:"activeCharacterID"`
	StationID         int `json:"stationID"`
}

func (m *databaseModel) GetCharacter(characterID int) (*Character, error) {
	const query = `SELECT characterName, owner, userID, refreshToken FROM characters WHERE characterID = ?`

	var refresh sql.NullString
	c := &Character{ID: characterID}
	err := m.db.QueryRow(query, characterID).Scan(&c.Name, &c.OwnerHash, &c.UserID, &refresh)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	c.RefreshToken = refresh.String
	return c, nil
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
		if err = rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		result[c.ID] = c
	}
	return result, rows.Err()
}

func (m *databaseModel) SaveUserStation(userID, stationID int) error {
	const query = `UPDATE users SET stationID = ? WHERE id = ?`
	_, err := m.db.Exec(query, stationID, userID)
	return err
}
