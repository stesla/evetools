package esd

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

var eveDB *sql.DB

var ErrNotFound = errors.New("Not Found")

func Initialize(dbfilename string) (err error) {
	eveDB, err = sql.Open("sqlite3", dbfilename)
	return
}

type MarketGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parentID"`
	Groups      []int  `json:"groups,omitempty"`
	Types       []int  `json:"types,omitempty"`
}

func GetMarketGroups() (map[int]*MarketGroup, error) {
	var query = `SELECT marketGroupID, marketGroupName, description, parentGroupID
			     FROM invMarketGroups`

	rows, err := eveDB.Query(query)
	if err != nil {
		return nil, err
	}
	result := map[int]*MarketGroup{}
	for rows.Next() {
		var id int
		var name, desc sql.NullString
		var pid sql.NullInt32
		if err := rows.Scan(&id, &name, &desc, &pid); err != nil {
			return nil, err
		}
		if name.Valid {
			result[id] = &MarketGroup{
				ID:          id,
				Name:        name.String,
				Description: desc.String,
			}
			if pid.Valid {
				var i = int(pid.Int32)
				result[id].ParentID = &i
			}
		}
	}
	return result, rows.Err()
}

type MarketType struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,-"`
	GroupID     int    `json:"groupID"`
}

func GetMarketTypes() (map[int]*MarketType, error) {
	var query = `SELECT typeID, typeName, description, marketGroupID FROM invTypes
				 WHERE published=1
				   AND marketGroupID IS NOT NULL`
	rows, err := eveDB.Query(query)
	if err != nil {
		return nil, err
	}
	result := map[int]*MarketType{}
	for rows.Next() {
		var id, groupid int
		var name, description sql.NullString
		if err := rows.Scan(&id, &name, &description, &groupid); err != nil {
			return nil, err
		}
		if name.Valid {
			result[id] = &MarketType{
				ID:          id,
				Name:        name.String,
				Description: description.String,
				GroupID:     groupid,
			}
		}
	}
	return result, rows.Err()
}

func SearchTypesByName(filter string) ([]int, error) {
	var query = `SELECT typeID FROM invTypes 
			       WHERE published=1
				     AND marketGroupID IS NOT NULL
			         AND typeName LIKE ?
                   ORDER BY typeName ASC`

	rows, err := eveDB.Query(query, "%"+filter+"%")
	if err != nil {
		return nil, err
	}
	result := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}
