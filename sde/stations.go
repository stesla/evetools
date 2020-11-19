package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type Station struct {
	ID       int    `yaml:"stationID" json:"id"`
	Name     string `yaml:"stationName" json:"name"`
	CorpID   int    `yaml:"corporationID" json:"corp_id"`
	RegionID int    `yaml:"regionID" json:"region_id"`
	SystemID int    `yaml:"solarSystemID" json:"system_id"`
}

func LoadStations() (map[int]Station, error) {
	input, err := os.Open(path.Join(sdeDir, "bsd", "staStations.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error opening staStations.yaml: %v", err)
	}
	defer input.Close()
	var yamlStations []Station
	err = yaml.NewDecoder(input).Decode(&yamlStations)

	var stations = map[int]Station{}
	for _, s := range yamlStations {
		stations[s.ID] = s
	}
	return stations, err
}
