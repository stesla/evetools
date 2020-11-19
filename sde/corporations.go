package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

type Corporation struct {
	ID        int    `json:"id"`
	Name      string `json:"string"`
	FactionID int    `json:"faction_id"`
}

type sdeCorporation struct {
	FactionID int `yaml:"factionID"`
	Name      struct {
		English string `yaml:"en"`
	} `yaml:"nameID"`
}

func LoadCorporations() (map[int]Corporation, error) {
	input, err := os.Open(path.Join(sdeDir, "fsd", "npcCorporations.yaml"))
	if err != nil {
		return nil, fmt.Errorf("error opening npcCorporations.yaml: %v", err)
	}
	defer input.Close()

	var sdeCorps map[int]sdeCorporation
	err = yaml.NewDecoder(input).Decode(&sdeCorps)

	result := map[int]Corporation{}
	for id, yc := range sdeCorps {
		if yc.FactionID == 0 {
			continue
		}
		jc := Corporation{
			ID:        id,
			Name:      yc.Name.English,
			FactionID: yc.FactionID,
		}
		result[id] = jc
	}

	return result, err
}
