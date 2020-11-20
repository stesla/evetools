package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type sdeStation struct {
	ID       int    `yaml:"stationID"`
	Name     string `yaml:"stationName"`
	CorpID   int    `yaml:"corporationID"`
	RegionID int    `yaml:"regionID"`
	SystemID int    `yaml:"solarSystemID"`
}

func buildStations(dir, pkg, file string) error {
	input, err := os.Open(path.Join(dir, "bsd", "staStations.yaml"))
	if err != nil {
		return fmt.Errorf("error opening staStations.yaml: %v", err)
	}
	defer input.Close()

	var yamlStations []sdeStation
	err = yaml.NewDecoder(input).Decode(&yamlStations)

	stations := map[int]sdeStation{}
	for _, s := range yamlStations {
		stations[s.ID] = s
	}

	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()

	t := template.Must(template.New("stations").Parse(stationsTemplate))
	return t.Execute(output, map[string]interface{}{
		"Package":  pkg,
		"Stations": stations,
	})

}

const stationsTemplate = `package {{ .Package }}

var Stations = map[int]*Station{ {{ range $id, $s := .Stations }}
  {{ $id }}: &Station{
	ID: {{ $id }},
	Name: "{{ $s.Name }}",
	CorpID: {{ $s.CorpID }},
	RegionID: {{ $s.RegionID }},
	SystemID: {{ $s.SystemID }},
  },{{ end }}
}`
