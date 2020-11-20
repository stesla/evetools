package main

import (
	"os"
	"path"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

type sdeMarketGroup struct {
	ParentID int `yaml:"parentGroupID"`

	Name struct {
		English string `yaml:"en"`
	} `yaml:"nameID"`

	Description struct {
		English string `yaml:"en"`
	} `yaml:"descriptionID"`
}

func buildMarketGroups(dir, pkg, file string) error {
	input, err := os.Open(path.Join(dir, "fsd", "marketGroups.yaml"))
	if err != nil {
		return err
	}
	defer input.Close()

	var sdeGroups map[int]sdeMarketGroup
	if err := yaml.NewDecoder(input).Decode(&sdeGroups); err != nil {
		return err
	}

	output, err := os.Create(file)
	if err != nil {
		return err
	}
	defer output.Close()

	t := template.Must(template.New("marketGroups").Parse(marketGroupsTemplate))
	return t.Execute(output, map[string]interface{}{
		"Package":      pkg,
		"MarketGroups": sdeGroups,
	})
}

const marketGroupsTemplate = `package {{ .Package }}

var MarketGroups = map[int]*MarketGroup{ {{ range $id, $g := .MarketGroups }}
  {{ $id }}: &MarketGroup{
	ID: {{ $id }},
	ParentID: {{ $g.ParentID }},
	Name: ` + "`{{ $g.Name.English }}`" + `,
	Description: ` + "`{{ $g.Description.English }}`" + `,
  },{{ end }}
}`
