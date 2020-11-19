package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type MarketType struct {
	ID            int    `json:"id"`
	MarketGroupID int    `json:"market_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
}

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

func LoadTypes() (map[int]*MarketType, error) {
	input, err := os.Open(path.Join(sdeDir, "fsd", "typeIDs.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error opening typeIDS.yaml: %v", err)
	}
	defer input.Close()
	var yamlTypes map[int]sdeMarketType
	if err := yaml.NewDecoder(input).Decode(&yamlTypes); err != nil {
		return nil, fmt.Errorf("error decoding typeIDs.yaml: %v", err)
	}

	var jsonTypes = map[int]*MarketType{}
	for id, yt := range yamlTypes {
		if yt.Published {
			jsonTypes[id] = &MarketType{
				ID:            id,
				MarketGroupID: yt.MarketGroupID,
				Name:          yt.Name.English,
				Description:   yt.Description.English,
			}
		}
	}
	return jsonTypes, err
}
