package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

var db *sql.DB

func initDatabase() (err error) {
	db, err = sql.Open("sqlite3", viper.GetString("model.database"))
	return
}

type MarketType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func GetMarketTypes(filter string) ([]*MarketType, error) {
	var query = `SELECT typeID, typeName FROM invTypes 
			       WHERE marketGroupID IS NOT NULL
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
