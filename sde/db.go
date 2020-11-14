package sde

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

var ErrNotFound = errors.New("Not Found")

type DB interface {
	GetStations(q string) (map[string]*Station, error)
	GetStationByID(stationID int) (*Station, error)
}

type staticDB struct {
	db *sql.DB
}

func Initialize(dbfilename string) (DB, error) {
	db, err := sql.Open("sqlite3", dbfilename)
	return &staticDB{db: db}, err
}

type Region struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type System struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Station struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Region Region `json:"region"`
	System System `json:"system"`
}

func (s *staticDB) GetStations(q string) (map[string]*Station, error) {
	return s.getStations("WHERE stationName LIKE ?", "%"+q+"%")
}

func (s *staticDB) getStations(clause string, args ...interface{}) (map[string]*Station, error) {
	var query = `SELECT s.stationID, s.stationName, s.solarSystemID, ss.solarSystemName, s.regionID, r.regionName
				   FROM staStations AS s
				   JOIN mapSolarSystems AS ss ON s.solarSystemID = ss.solarSystemID
				   JOIN mapRegions AS r ON s.regionID = r.regionID ` + clause
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	result := map[string]*Station{}
	for rows.Next() {
		var s Station
		if err := rows.Scan(&s.ID, &s.Name, &s.System.ID, &s.System.Name, &s.Region.ID, &s.Region.Name); err != nil {
			return nil, err
		}
		result[s.Name] = &s
	}
	return result, rows.Err()
}

func (s *staticDB) GetStationByID(stationID int) (*Station, error) {
	stations, err := s.getStations("WHERE stationID = ?", stationID)
	if err != nil {
		return nil, err
	}
	for _, v := range stations {
		return v, nil
	}
	return nil, ErrNotFound
}
