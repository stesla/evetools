package sde

import (
	"database/sql"
	"errors"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var ErrNotFound = errors.New("Not Found")

type DB interface {
	GetMarketGroups() (map[int]*MarketGroup, error)
	GetMarketTypes() (map[int]*MarketType, error)
	GetStations(q string) (map[string]*Station, error)
	GetStationByID(stationID int) (*Station, error)
	SearchTypesByName(filter string) ([]int, error)
}

type staticDB struct {
	db *sql.DB
}

func Initialize(dbfilename string) (DB, error) {
	db, err := sql.Open("sqlite3", dbfilename)
	return &staticDB{db: db}, err
}

type MarketGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parentID"`
	Groups      []int  `json:"groups,omitempty"`
	Types       []int  `json:"types,omitempty"`
}

func (s *staticDB) GetMarketGroups() (map[int]*MarketGroup, error) {
	var query = `SELECT marketGroupID, marketGroupName, description, parentGroupID
			     FROM invMarketGroups`

	rows, err := s.db.Query(query)
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

func (s *staticDB) GetMarketTypes() (map[int]*MarketType, error) {
	var query = `SELECT typeID, typeName, description, marketGroupID FROM invTypes
				 WHERE published=1
				   AND marketGroupID IS NOT NULL`
	rows, err := s.db.Query(query)
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

type Station struct {
	ID       int    `json:"id"`
	RegionID int    `json:"regionID"`
	Name     string `json:"name"`
}

func (s *staticDB) GetStations(q string) (map[string]*Station, error) {
	const query = `SELECT stationID, regionID, stationName FROM staStations
				 WHERE stationName LIKE ?`
	rows, err := s.db.Query(query, "%"+q+"%")
	if err != nil {
		return nil, err
	}
	result := map[string]*Station{}
	for rows.Next() {
		var s Station
		if err := rows.Scan(&s.ID, &s.RegionID, &s.Name); err != nil {
			return nil, err
		}
		result[s.Name] = &s
	}
	return result, rows.Err()
}

func (s *staticDB) GetStationByID(stationID int) (*Station, error) {
	const query = `SELECT regionID, stationName FROM staStations
				 WHERE stationID = ?`
	sta := &Station{ID: stationID}
	log.Println(stationID)
	err := s.db.QueryRow(query, stationID).Scan(&sta.RegionID, &sta.Name)
	return sta, err
}

func (s *staticDB) SearchTypesByName(filter string) ([]int, error) {
	var query = `SELECT typeID FROM invTypes 
			       WHERE published=1
				     AND marketGroupID IS NOT NULL
			         AND typeName LIKE ?
                   ORDER BY typeName ASC`

	rows, err := s.db.Query(query, "%"+filter+"%")
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
