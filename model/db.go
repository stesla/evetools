package model

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var ErrNotFound = errors.New("Not Found")

func Initialize(dbfilename string) (err error) {
	db, err = sql.Open("sqlite3", dbfilename)
	return
}

type Type struct {
	ID       int  `json:"id"`
	Favorite bool `json:"favorite"`
}

func GetType(typeID int) (t Type, err error) {
	const query = `SELECT favorite FROM types WHERE typeID = ?`

	t.ID = typeID
	err = db.QueryRow(query, typeID).Scan(&t.Favorite)
	if err == sql.ErrNoRows {
		err = ErrNotFound
	}
	return
}

func FavoriteTypes() ([]Type, error) {
	rows, err := db.Query("SELECT typeID, favorite FROM types WHERE favorite")
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

func SetFavorite(typeID int, val bool) error {
	t, err := GetType(typeID)
	if err == ErrNotFound {
		// if it is not a favorite, it does not need a record created
		if val {
			_, err = db.Exec("INSERT INTO types (typeID, favorite) VALUES (?, ?)", typeID, val)
		}
	} else if err == nil {
		// only do a query if we're actually changing the value
		if val != t.Favorite {
			_, err = db.Exec("UPDATE types SET favorite = ? WHERE typeID = ?", val, typeID)
		}
	}
	return err
}
