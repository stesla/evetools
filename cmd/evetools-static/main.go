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

	var err error
	var types map[int]*sde.MarketType

	err = sde.Initialize(*sdeDir)
	if err != nil {
		die(err)
	}

	if *convertTypes || *convertGroups {
		types, err = sde.LoadTypes()
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
		groups, root, err := sde.LoadGroups(types)
		if err != nil {
			die(fmt.Errorf("error loading groups: %v", err))
		}

		err = saveGroups(*outDir, groups, root)
		if err != nil {
			die(fmt.Errorf("error saving groups: %v", err))
		}
	}

	if *convertStations {
		stations, err := sde.LoadStations()
		if err != nil {
			die(fmt.Errorf("error loading stations: %v", err))
		}

		err = saveStations(*outDir, stations)
		if err != nil {
			die(fmt.Errorf("error saving stations: %v", err))
		}
	}

	if *convertSystems {
		systems, err := sde.LoadSystems()
		if err != nil {
			die(fmt.Errorf("error loading systems: %v", err))
		}

		err = saveSystems(*outDir, systems)
		if err != nil {
			die(fmt.Errorf("error saving systems: %v", err))
		}
	}

	if *convertCorps {
		corps, err := sde.LoadCorporations()
		if err != nil {
			die(fmt.Errorf("error loading corps: %v", err))
		}

		err = saveCorps(*outDir, corps)
		if err != nil {
			die(fmt.Errorf("error saving corps: %v", err))
		}
	}
}

func saveTypes(dir string, jsonTypes map[int]*sde.MarketType) error {
	output, err := os.Create(path.Join(dir, "types.json"))
	if err != nil {
		return fmt.Errorf("error opening types.json: %v", err)
	}
	defer output.Close()
	return json.NewEncoder(output).Encode(&jsonTypes)
}

func saveGroups(dir string, jsonGroups map[int]*sde.MarketGroup, root []int) error {
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
