package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

var stations map[int]*Station

type Station struct {
	ID       int    `yaml:"stationID" json:"id"`
	Name     string `yaml:"stationName" json:"name"`
	CorpID   int    `yaml:"corporationID" json:"corp_id"`
	RegionID int    `yaml:"regionID" json:"region_id"`
	SystemID int    `yaml:"solarSystemID" json:"system_id"`
}

func loadStations(dir string) error {
	input, err := os.Open(path.Join(dir, "bsd", "staStations.yaml"))
	if err != nil {
		return fmt.Errorf("error opening staStations.yaml: %v", err)
	}
	defer input.Close()
	var yamlStations []*Station
	err = yaml.NewDecoder(input).Decode(&yamlStations)

	stations = map[int]*Station{}
	for _, s := range yamlStations {
		stations[s.ID] = s
	}
	return err
}

func GetStations() map[int]*Station {
	return stations
}

func GetStation(id int) (s *Station, found bool) {
	s, found = stations[id]
	return
}
