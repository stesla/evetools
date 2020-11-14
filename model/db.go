package model

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	AssociateWithUser(userID, characterID int, characterName, owner, refreshToken string) error
	GetFavoriteTypes(userID int) ([]int, error)
	FindOrCreateUserForCharacter(characterID int, characterName, owner, refreshToken string) (*User, error)
	GetCharacter(int) (*Character, error)
	GetCharactersForUser(int) (map[int]*Character, error)
	IsFavorite(userID, typeID int) (bool, error)
	RemoveUserAssociation(characterID int) error
	SaveUserStation(userID, stationID int) error
	SetActiveCharacter(userID, characterID int) error
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

func (m *databaseModel) AssociateWithUser(userID, characterID int, characterName, owner, refreshToken string) (err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	var uid sql.NullInt32
	query := "SELECT userID FROM characters WHERE characterID = ? AND userID IS NOT NULL"
	if err = tx.QueryRow(query, characterID).Scan(&uid); err == sql.ErrNoRows {
		query = `INSERT INTO characters
			   (userID, characterID, characterName, owner, refreshToken)
			   VALUES (?, ?, ?, ?, ?)`
		_, err = tx.Exec(query, userID, characterID, characterName, owner, refreshToken)
	} else if err != nil {
		return
	} else if !uid.Valid {
		query = `UPDATE characters
				 SET userID = ?, characterName = ?, owner = ?, refreshToken = ?
				 WHERE characterID = ?`
		_, err = tx.Exec(query, userID, characterName, owner, refreshToken, characterID)
	} else if int(uid.Int32) != userID {
		log.Println(uid.Int32, userID, characterID, characterName)
		err = errors.New("character already associated with another user")
	}
	return
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
	Owner        string `json:"-"`
	UserID       int    `json:"-"`
	RefreshToken string `json:"-"`
}

type User struct {
	ID                int `json:"id"`
	ActiveCharacterID int `json:"activeCharacterID"`
	StationID         int `json:"stationID"`
}

func (m *databaseModel) FindOrCreateUserForCharacter(characterID int, characterName, owner, refreshToken string) (*User, error) {
	const getUserID = `SELECT c.userID, u.stationID, u.activeCharacterID
					   FROM characters AS c JOIN users AS u ON c.userID = u.ID
					   WHERE characterID = ? AND owner = ?`
	const createCharacter = `INSERT INTO characters 
							 (characterID, characterName, owner, userID, refreshToken)
							 VALUES (?, ?, ?, ?, ?)`
	const createUser = `INSERT INTO users (activeCharacterID) VALUES (?)`

	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}

	var userID int
	var stationID int = 60003760 // Jita IV - Moon 4 - Caldari Navy Assembly Plant
	var activeID int
	if err = tx.QueryRow(getUserID, characterID, owner).Scan(&userID, &stationID, &activeID); err == sql.ErrNoRows {
		activeID = characterID
		var r sql.Result
		r, err = tx.Exec(createUser, characterID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		var id int64
		id, err = r.LastInsertId()
		userID = int(id)

		_, err = tx.Exec(createCharacter, characterID, characterName, owner, userID, refreshToken)
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
		ActiveCharacterID: activeID,
		StationID:         stationID,
	}, tx.Commit()
}
func (m *databaseModel) GetCharacter(characterID int) (*Character, error) {
	const query = `SELECT characterName, owner, userID, refreshToken FROM characters WHERE characterID = ?`

	var refresh sql.NullString
	c := &Character{ID: characterID}
	err := m.db.QueryRow(query, characterID).Scan(&c.Name, &c.Owner, &c.UserID, &refresh)
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

func (m *databaseModel) RemoveUserAssociation(characterID int) error {
	const query = `DELETE FROM characters WHERE characterID = ?`
	_, err := m.db.Exec(query, characterID)
	return err
}

func (m *databaseModel) SaveUserStation(userID, stationID int) error {
	const query = `UPDATE users SET stationID = ? WHERE id = ?`
	_, err := m.db.Exec(query, stationID, userID)
	return err
}

func (m *databaseModel) SetActiveCharacter(userID, characterID int) error {
	const query = `UPDATE users SET activeCharacterID = ? WHERE id = ?`
	_, err := m.db.Exec(query, characterID, userID)
	return err
}
