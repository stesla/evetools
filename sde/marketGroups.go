package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

var marketGroups map[int]*MarketGroup
var roots []int

type MarketGroup struct {
	ID          int    `json:"id"`
	ParentID    int    `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Groups []int `json:"-"`
	Types  []int `json:"-"`
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

func loadGroups(dir string, types map[int]MarketType) error {
	input, err := os.Open(path.Join(dir, "fsd", "marketGroups.yaml"))
	if err != nil {
		return fmt.Errorf("error opening marketGroups.yaml: %v", err)
	}
	defer input.Close()
	var yamlGroups map[int]sdeMarketGroup
	if err := yaml.NewDecoder(input).Decode(&yamlGroups); err != nil {
		return fmt.Errorf("error decoding marketGroups.yaml: %v", err)
	}

	marketGroups = map[int]*MarketGroup{}
	for id, yg := range yamlGroups {
		marketGroups[id] = &MarketGroup{
			ID:          id,
			ParentID:    yg.ParentID,
			Name:        yg.Name.English,
			Description: yg.Description.English,
		}
	}

	for id, jg := range marketGroups {
		if jg.ParentID == 0 {
			roots = append(roots, id)
			continue
		}
		pg := marketGroups[jg.ParentID]
		pg.Groups = append(pg.Groups, id)
	}

	for id, jt := range types {
		if pg, found := marketGroups[jt.MarketGroupID]; found {
			pg.Types = append(pg.Types, id)
		}
	}

	return nil
}

func GetMarketGroups() (map[int]*MarketGroup, []int) {
	return marketGroups, roots
}

func GetMarketGroup(id int) (g *MarketGroup, found bool) {
	g, found = marketGroups[id]
	return
}
