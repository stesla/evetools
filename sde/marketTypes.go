package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

var marketTypes map[int]*MarketType

type MarketType struct {
	ID            int    `json:"id"`
	MarketGroupID int    `json:"market_group_id"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
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

func loadTypes(dir string) error {
	input, err := os.Open(path.Join(dir, "fsd", "typeIDs.yaml"))
	if err != nil {
		return fmt.Errorf("error loading market types: %v", err)
	}
	defer input.Close()
	var yamlTypes map[int]sdeMarketType
	if err := yaml.NewDecoder(input).Decode(&yamlTypes); err != nil {
		return fmt.Errorf("error loading market types: %v", err)
	}

	marketTypes = map[int]*MarketType{}
	for id, yt := range yamlTypes {
		if yt.Published {
			marketTypes[id] = &MarketType{
				ID:            id,
				MarketGroupID: yt.MarketGroupID,
				Name:          yt.Name.English,
				Description:   yt.Description.English,
			}
		}
	}
	return nil
}

func GetMarketTypes() (out map[int]*MarketType) {
	out = make(map[int]*MarketType, len(marketTypes))
	for k, v := range marketTypes {
		var t MarketType = *v
		out[k] = &t
	}
	return
}

func GetMarketType(id int) (*MarketType, bool) {
	v, found := marketTypes[id]
	if !found {
		return nil, found
	}
	var t MarketType = *v
	return &t, found
}
