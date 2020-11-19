package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/stesla/evetools/sde"
)

var (
	sdeDir          = flag.String("in", "./data/sde", "directory in which to look for the SDE YAML files")
	outDir          = flag.String("out", "./public/data", "directory into which JSON files should be placed")
	convertTypes    = flag.Bool("types", false, "convert market types")
	convertGroups   = flag.Bool("groups", false, "convert market groups")
	convertStations = flag.Bool("stations", false, "convert stations")
	convertSystems  = flag.Bool("systems", false, "convert systems")
	convertCorps    = flag.Bool("corps", false, "convert corps")
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

	err := sde.Initialize(*sdeDir)
	if err != nil {
		die(err)
	}

	if *convertTypes {
		types := sde.GetMarketTypes()
		err = saveTypes(*outDir, types)
		if err != nil {
			die(fmt.Errorf("error saving types: %v", err))
		}
	}

	if *convertGroups {
		groups, root := sde.GetMarketGroups()
		err = saveGroups(*outDir, groups, root)
		if err != nil {
			die(fmt.Errorf("error saving groups: %v", err))
		}
	}

	if *convertStations {
		stations := sde.GetStations()
		err = saveStations(*outDir, stations)
		if err != nil {
			die(fmt.Errorf("error saving stations: %v", err))
		}
	}

	if *convertSystems {
		systems := sde.GetSolarSystems()
		err = saveSystems(*outDir, systems)
		if err != nil {
			die(fmt.Errorf("error saving systems: %v", err))
		}
	}

	if *convertCorps {
		corps := sde.GetCorporations()
		err = saveCorps(*outDir, corps)
		if err != nil {
			die(fmt.Errorf("error saving corps: %v", err))
		}
	}
}

func saveTypes(dir string, marketTypes map[int]sde.MarketType) error {
	output, err := os.Create(path.Join(dir, "types.json"))
	if err != nil {
		return fmt.Errorf("error opening types.json: %v", err)
	}
	defer output.Close()

	outTypes := make(map[int]sde.MarketType, len(marketTypes))
	for i, t := range marketTypes {
		var o sde.MarketType
		o.ID = t.ID
		o.MarketGroupID = t.MarketGroupID
		o.Name = t.Name
		outTypes[i] = o
	}
	return json.NewEncoder(output).Encode(&outTypes)
}

func saveGroups(dir string, jsonGroups map[int]sde.MarketGroup, root []int) error {
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

func saveStations(dir string, stations map[int]sde.Station) error {
	output, err := os.Create(path.Join(dir, "stations.json"))
	if err != nil {
		return fmt.Errorf("error opening stations.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&stations)
}

func saveSystems(dir string, systems map[int]sde.SolarSystem) error {
	output, err := os.Create(path.Join(dir, "systems.json"))
	if err != nil {
		return fmt.Errorf("error opening systems.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&systems)
}

func saveCorps(dir string, corps map[int]sde.Corporation) error {
	output, err := os.Create(path.Join(dir, "corporations.json"))
	if err != nil {
		return fmt.Errorf("error opening corporations.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&corps)
}
