package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type sdeCorporation struct {
	FactionID int `yaml:"factionID"`
	Name      struct {
		English string `yaml:"en"`
	} `yaml:"nameID"`
}

func buildCorporations(dir, pkg, file string) error {
	input, err := os.Open(path.Join(dir, "fsd", "npcCorporations.yaml"))
	if err != nil {
		return fmt.Errorf("error opening npcCorporations.yaml: %v", err)
	}
	defer input.Close()

	var sdeCorps map[int]sdeCorporation
	err = yaml.NewDecoder(input).Decode(&sdeCorps)
	if err != nil {
		return err
	}

	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()

	t := template.Must(template.New("corporations").Parse(corporationsTemplate))
	return t.Execute(output, map[string]interface{}{
		"Package":      pkg,
		"Corporations": sdeCorps,
	})
}

const corporationsTemplate = `package {{ .Package }}

var Corporations = map[int]*Corporation{ {{ range $id, $corp := .Corporations }}
  {{ $id }}: &Corporation{ID: {{ $id }}, Name: ` + "`{{ $corp.Name.English }}`" + `, FactionID: {{ $corp.FactionID }} },{{ end }}
}
`
