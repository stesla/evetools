package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type SolarSystem struct {
	ID       int    `yaml:"solarSystemID"`
	Name     string `yaml:"name"`
	RegionID int    `yaml:"regionID"`
}

func buildSolarSystems(dir, pkg, file string) error {
	glob := path.Join(dir, "fsd", "universe", "*", "*", "region.staticdata")

	paths, err := filepath.Glob(glob)
	if err != nil {
		return fmt.Errorf("error listing regions: %v", err)
	}

	solarSystems := map[int]SolarSystem{}
	for _, p := range paths {
		r, err := loadRegion(p)
		if err != nil {
			return fmt.Errorf("error loading region: %v", err)
		}
		for _, s := range r {
			solarSystems[s.ID] = s
		}
	}

	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()

	t := template.Must(template.New("solarSystems").Parse(solarSystemsTemplate))
	return t.Execute(output, map[string]interface{}{
		"Package":      pkg,
		"SolarSystems": solarSystems,
	})
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

const solarSystemsTemplate = `package {{ .Package }}

var SolarSystems = map[int]*SolarSystem{ {{ range $id, $s := .SolarSystems }}
  {{ $id }}: &SolarSystem{
	ID: {{ $id }},
	Name: "{{ $s.Name }}",
	RegionID: {{ $s.RegionID }},
  },{{ end }}
}`
