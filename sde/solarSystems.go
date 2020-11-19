package sde

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type SolarSystem struct {
	ID       int    `json:"id" yaml:"solarSystemID"`
	Name     string `json:"name"`
	RegionID int    `json:"region_id" yaml:"regionID"`
}

func LoadSystems() (map[int]SolarSystem, error) {
	glob := path.Join(sdeDir, "fsd", "universe", "*", "*", "region.staticdata")

	paths, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("error listing regions: %v", err)
	}

	var systems = map[int]SolarSystem{}
	for _, p := range paths {
		r, err := loadRegion(p)
		if err != nil {
			return nil, fmt.Errorf("error loading region: %v", err)
		}
		for _, s := range r {
			systems[s.ID] = s
		}
	}
	return systems, nil
}

func loadRegion(filename string) ([]SolarSystem, error) {
	input, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening region data: %v", err)
	}
	defer input.Close()

	var region SolarSystem // just using this to capture the region ID
	err = yaml.NewDecoder(input).Decode(&region)
	if err != nil {
		return nil, fmt.Errorf("error decoding region data: %v", err)
	}

	dir, _ := filepath.Split(filename)
	glob := path.Join(dir, "*", "*", "solarsystem.staticdata")
	paths, err := filepath.Glob(glob)
	if err != nil {
		return nil, fmt.Errorf("error listing systems; %v", err)
	}
	var result []SolarSystem
	for _, p := range paths {
		s, err := loadSolarSystem(p)
		if err != nil {
			return nil, fmt.Errorf("error loading solar system: %v", err)
		}
		s.RegionID = region.RegionID
		result = append(result, s)
	}
	return result, nil
}

func loadSolarSystem(filename string) (SolarSystem, error) {
	input, err := os.Open(filename)
	if err != nil {
		return SolarSystem{}, fmt.Errorf("error opening solar system data: %v", err)
	}
	defer input.Close()

	dir, _ := filepath.Split(filename)
	ss := SolarSystem{
		Name: filepath.Base(dir),
	}
	err = yaml.NewDecoder(input).Decode(&ss)
	return ss, err
}
