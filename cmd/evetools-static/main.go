package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

var (
	sdeDir          = flag.String("in", "./data/sde", "directory in which to look for the SDE YAML files")
	outDir          = flag.String("out", "./public/data", "directory into which JSON files should be placed")
	convertTypes    = flag.Bool("types", false, "convert market types")
	convertGroups   = flag.Bool("groups", false, "convert market groups")
	convertStations = flag.Bool("stations", false, "convert stations")
)

func usage() error {
	program := filepath.Base(os.Args[0])
	return fmt.Errorf("USAGE: %s [-in DIR] [-out DIR]", program)
}

func die(err error) {
	fmt.Fprintf(os.Stderr, err.Error())
	os.Exit(1)
}

func main() {
	flag.Parse()

	if (*sdeDir) == "" || (*outDir) == "" {
		die(usage())
	}

	var err error
	var types map[int]*JsonType

	if *convertTypes || *convertGroups {
		types, err = loadTypes(*sdeDir)
		if err != nil {
			die(fmt.Errorf("error loading types: %v", err))
		}
	}

	if *convertTypes {
		err = saveTypes(*outDir, types)
		if err != nil {
			die(fmt.Errorf("error saving types: %v", err))
		}
	}

	if *convertGroups {
		groups, root, err := loadGroups(types)
		if err != nil {
			die(fmt.Errorf("error loading groups: %v", err))
		}

		err = saveGroups(*outDir, groups, root)
		if err != nil {
			die(fmt.Errorf("error saving groups: %v", err))
		}
	}

	if *convertStations {
		stations, err := loadStations(*sdeDir)
		if err != nil {
			die(fmt.Errorf("error loading stations: %v", err))
		}

		err = saveStations(*outDir, stations)
		if err != nil {
			die(fmt.Errorf("error saving stations: %v", err))
		}
	}
}

type YamlType struct {
	MarketGroupID int  `yaml:"marketGroupID"`
	Published     bool `yaml:"published"`

	Name struct {
		English string `yaml:"en"`
	} `yaml:"name"`

	Description struct {
		English string `yaml:"en"`
	} `yaml:"description"`
}

type JsonType struct {
	ID            int    `json:"id"`
	MarketGroupID int    `json:"market_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

func loadTypes(dir string) (map[int]*JsonType, error) {
	input, err := os.Open(path.Join(dir, "fsd", "typeIDs.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error opening typeIDS.yaml: %v", err)
	}
	defer input.Close()
	var yamlTypes map[int]YamlType
	if err := yaml.NewDecoder(input).Decode(&yamlTypes); err != nil {
		return nil, fmt.Errorf("error decoding typeIDs.yaml: %v", err)
	}

	var jsonTypes = map[int]*JsonType{}
	for id, yt := range yamlTypes {
		if yt.Published {
			jsonTypes[id] = &JsonType{
				ID:            id,
				MarketGroupID: yt.MarketGroupID,
				Name:          yt.Name.English,
				Description:   yt.Description.English,
			}
		}
	}
	return jsonTypes, err
}

func saveTypes(dir string, jsonTypes map[int]*JsonType) error {
	output, err := os.Create(path.Join(dir, "types.json"))
	if err != nil {
		return fmt.Errorf("error opening types.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&jsonTypes)
}

type YamlGroup struct {
	ParentID int `yaml:"parentGroupID"`

	Name struct {
		English string `yaml:"en"`
	} `yaml:"nameID"`

	Description struct {
		English string `yaml:"en"`
	} `yaml:"descriptionID"`
}

type JsonGroup struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Groups []int `json:"groups"`
	Types  []int `json:"types"`
}

func loadGroups(types map[int]*JsonType) (map[int]*JsonGroup, []int, error) {
	input, err := os.Open(path.Join(*sdeDir, "fsd", "marketGroups.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("error opening marketGroups.yaml: %v", err)
	}
	defer input.Close()
	var yamlGroups map[int]YamlGroup
	if err := yaml.NewDecoder(input).Decode(&yamlGroups); err != nil {
		return nil, nil, fmt.Errorf("error decoding marketGroups.yaml: %v", err)
	}

	var jsonGroups = map[int]*JsonGroup{}
	for id, yg := range yamlGroups {
		jsonGroups[id] = &JsonGroup{
			ID:          id,
			ParentID:    yg.ParentID,
			Name:        yg.Name.English,
			Description: yg.Description.English,
		}
	}

	var root []int
	for id, jg := range jsonGroups {
		if jg.ParentID == 0 {
			root = append(root, id)
			continue
		}
		pg := jsonGroups[jg.ParentID]
		pg.Groups = append(pg.Groups, id)
	}

	for id, jt := range types {
		if pg, found := jsonGroups[jt.MarketGroupID]; found {
			pg.Types = append(pg.Types, id)
		}
	}

	return jsonGroups, root, nil
}

func saveGroups(dir string, jsonGroups map[int]*JsonGroup, root []int) error {
	output, err := os.Create(path.Join(dir, "marketGroups.json"))
	if err != nil {
		return fmt.Errorf("error opening marketGroups.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(map[string]interface{}{
		"groups": jsonGroups,
		"root":   root,
	})
}

type Station struct {
	ID       int    `yaml:"stationID" json:"id"`
	Name     string `yaml:"stationName" json:"name"`
	RegionID int    `yaml:"regionID" json:"region_id"`
	SystemID int    `yaml:"solarSystemID" json:"system_id"`
}

func loadStations(dir string) (map[string]Station, error) {
	input, err := os.Open(path.Join(dir, "bsd", "staStations.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error opening staStations.yaml: %v", err)
	}
	defer input.Close()
	var yamlStations []Station
	if err := yaml.NewDecoder(input).Decode(&yamlStations); err != nil {
		return nil, fmt.Errorf("error decoding staStations.yaml: %v", err)
	}

	var stations = map[string]Station{}
	for _, station := range yamlStations {
		stations[station.Name] = station
	}
	return stations, nil
}

func saveStations(dir string, stations map[string]Station) error {
	output, err := os.Create(path.Join(dir, "stations.json"))
	if err != nil {
		return fmt.Errorf("error opening stations.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&stations)
}
