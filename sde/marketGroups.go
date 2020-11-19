package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type MarketGroup struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Groups []int `json:"groups"`
	Types  []int `json:"types"`
}

type sdeMarketGroup struct {
	ParentID int `yaml:"parentGroupID"`

	Name struct {
		English string `yaml:"en"`
	} `yaml:"nameID"`

	Description struct {
		English string `yaml:"en"`
	} `yaml:"descriptionID"`
}

func LoadGroups(types map[int]*MarketType) (map[int]*MarketGroup, []int, error) {
	input, err := os.Open(path.Join(sdeDir, "fsd", "marketGroups.yaml"))
	if err != nil {
		return nil, nil, fmt.Errorf("error opening marketGroups.yaml: %v", err)
	}
	defer input.Close()
	var yamlGroups map[int]sdeMarketGroup
	if err := yaml.NewDecoder(input).Decode(&yamlGroups); err != nil {
		return nil, nil, fmt.Errorf("error decoding marketGroups.yaml: %v", err)
	}

	var jsonGroups = map[int]*MarketGroup{}
	for id, yg := range yamlGroups {
		jsonGroups[id] = &MarketGroup{
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
