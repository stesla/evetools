package main

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

var db *sql.DB

var ErrNotFound = errors.New("Not Found")

func initDatabase() (err error) {
	db, err = sql.Open("sqlite3", viper.GetString("model.database"))
	return
}

type MarketType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,-"`
}

func GetMarketType(id int) (*MarketType, error) {
	var query = `SELECT typeName, description FROM invTypes WHERE typeID = ?`
	var name, desc sql.NullString
	err := db.QueryRow(query, id).Scan(&name, &desc)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &MarketType{
		ID:          id,
		Name:        name.String,
		Description: desc.String,
	}, nil
}

func GetMarketTypes(filter string) ([]*MarketType, error) {
	var query = `SELECT typeID, typeName FROM invTypes 
			       WHERE published=1
				     AND marketGroupID IS NOT NULL
			         AND typeName LIKE ?
                   ORDER BY typeName ASC`

	rows, err := db.Query(query, "%"+filter+"%")
	if err != nil {
		return nil, err
	}
	result := []*MarketType{}
	for rows.Next() {
		var id sql.NullInt32
		var name sql.NullString
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		if id.Valid && name.Valid {
			result = append(result, &MarketType{ID: int(id.Int32), Name: name.String})
		}
	}
	return result, rows.Err()
}
