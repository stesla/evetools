package sde

import (
	"fmt"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

var corporations map[int]*Corporation

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

func loadCorporations(dir string) error {
	input, err := os.Open(path.Join(dir, "fsd", "npcCorporations.yaml"))
	if err != nil {
		return fmt.Errorf("error opening npcCorporations.yaml: %v", err)
	}
	defer input.Close()

	var sdeCorps map[int]sdeCorporation
	err = yaml.NewDecoder(input).Decode(&sdeCorps)

	corporations = map[int]*Corporation{}
	for id, yc := range sdeCorps {
		if yc.FactionID == 0 {
			continue
		}
		jc := &Corporation{
			ID:        id,
			Name:      yc.Name.English,
			FactionID: yc.FactionID,
		}
		corporations[id] = jc
	}

	return err
}

func GetCorporations() map[int]*Corporation {
	return corporations
}

func GetCorporation(id int) (c *Corporation, found bool) {
	c, found = corporations[id]
	return
}
