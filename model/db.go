package model

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	GetType(int) (Type, error)
	FavoriteTypes() ([]Type, error)
	SetFavorite(int, bool) error
	FindOrCreateUserForCharacter(int, string, string) (*User, error)
	GetCharacter(int) (*Character, error)
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

type Type struct {
	ID       int  `json:"id"`
	Favorite bool `json:"favorite"`
}

func (m *databaseModel) GetType(typeID int) (t Type, err error) {
	const query = `SELECT favorite FROM types WHERE typeID = ?`

	t.ID = typeID
	err = m.db.QueryRow(query, typeID).Scan(&t.Favorite)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func (m *databaseModel) FavoriteTypes() ([]Type, error) {
	rows, err := m.db.Query("SELECT typeID, favorite FROM types WHERE favorite")
	if err != nil {
		return nil, err
	}
	result := []Type{}
	for rows.Next() {
		var t Type
		if err := rows.Scan(&t.ID, &t.Favorite); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (m *databaseModel) SetFavorite(typeID int, val bool) error {
	t, err := m.GetType(typeID)
	if err == ErrNotFound {
		// if it is not a favorite, it does not need a record created
		if val {
			_, err = m.db.Exec("INSERT INTO types (typeID, favorite) VALUES (?, ?)", typeID, val)
		}
	} else if err == nil {
		// only do a query if we're actually changing the value
		if val != t.Favorite {
			_, err = m.db.Exec("UPDATE types SET favorite = ? WHERE typeID = ?", val, typeID)
		}
	}
	return err
}

type Character struct {
	CharacterID   int    `json:"characterID"`
	CharacterName string `json:"characterName"`
	Owner         string `json:"-"`
	UserID        string `json:"-"`
}

type User struct {
	ID                int `json:"id"`
	ActiveCharacterID int `json:"activeCharacterID"`
}

func (m *databaseModel) FindOrCreateUserForCharacter(characterID int, characterName, owner string) (*User, error) {
	const getUserID = `SELECT userID FROM characters WHERE characterID = ? AND owner = ?`
	const createCharacter = `INSERT INTO characters (characterID, characterName, owner, userID) VALUES (?, ?, ?, ?)`
	const createUser = `INSERT INTO users (activeCharacterID) VALUES (?)`

	tx, err := m.db.Begin()
	if err != nil {
		return nil, err
	}

	var userID int
	if err = tx.QueryRow(getUserID, characterID, owner).Scan(&userID); err == sql.ErrNoRows {
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
	return &User{ID: userID, ActiveCharacterID: characterID}, tx.Commit()
}

func (m *databaseModel) GetCharacter(characterID int) (*Character, error) {
	const query = `SELECT characterName, owner, userID FROM characters WHERE characterID = ?`

	c := &Character{CharacterID: characterID}
	err := m.db.QueryRow(query, characterID).Scan(&c.CharacterName, &c.Owner, &c.UserID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return c, nil
}
