package main

import (
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type sdeMarketType struct {
	MarketGroupID int  `yaml:"marketGroupID"`
	Published     bool `yaml:"published"`

	Name struct {
		English string `yaml:"en"`
	} `yaml:"name"`

	Description struct {
		English string `yaml:"en"`
	} `yaml:"description"`
}

func buildMarketTypes(dir, pkg, file string) error {
	input, err := os.Open(path.Join(dir, "fsd", "typeIDs.yaml"))
	if err != nil {
		return err
	}
	defer input.Close()

	var sdeTypes map[int]sdeMarketType
	if err := yaml.NewDecoder(input).Decode(&sdeTypes); err != nil {
		return err
	}

	types := map[int]sdeMarketType{}
	for id, t := range sdeTypes {
		if t.Published && t.MarketGroupID > 0 {
			types[id] = t
		}
	}

	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()

	t := template.Must(template.New("marketTypes").Parse(marketTypesTemplate))
	return t.Execute(output, map[string]interface{}{
		"Package":     pkg,
		"MarketTypes": types,
	})

}

const marketTypesTemplate = `package {{ .Package }}

var MarketTypes = map[int]*MarketType{ {{ range $id, $t := .MarketTypes }}
  {{ $id }}: &MarketType{
	ID: {{ $id }},
	MarketGroupID: {{ $t.MarketGroupID }},
	Name: ` + "`{{ $t.Name.English }}`" + `,
	Description: ` + "`{{ $t.Description.English }}`" + `,
  },{{ end }}
}
`
