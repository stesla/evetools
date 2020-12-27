package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stesla/evetools/sde"
)

func (c *client) GetCharacterLocation(ctx context.Context, characterID int) (string, error) {
	url := fmt.Sprintf("/characters/%d/location/", characterID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var obj struct {
		SystemID    int  `json:"solar_system_id"`
		StationID   *int `json:"station_id"`
		StructureID *int `json:"structure_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		return "", err
	}

	system, found := sde.SolarSystems[obj.SystemID]
	if !found {
		return "", fmt.Errorf("System not found %d", obj.SystemID)
	}

	if obj.StationID != nil {
		station, found := sde.Stations[*obj.StationID]
		if !found {
			return fmt.Sprintf("%s - Unknown Station %d", system.Name, *obj.StationID), nil
		}
		return station.Name, nil
	} else if obj.StructureID != nil {
		return fmt.Sprintf("%s - Unknown Structure %d", system.Name, *obj.StructureID), nil
	}
	return system.Name, nil
}

func (c *client) GetCharacterName(characterID int) (string, error) {
	url := fmt.Sprintf("/characters/%d/", characterID)
	req, err := newESIRequest(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var character struct {
		Name string `json:"name"`
	}
	err = json.NewDecoder(resp.Body).Decode(&character)
	return character.Name, err
}

type Skill struct {
	ID          int `json:"skill_id"`
	ActiveLevel int `json:"active_skill_level"`
}

func (c *client) GetCharacterSkills(ctx context.Context, characterID int) ([]Skill, error) {
	url := fmt.Sprintf("/characters/%d/skills/", characterID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var obj struct {
		Skills []Skill `json:"skills"`
	}
	err = json.NewDecoder(resp.Body).Decode(&obj)
	return obj.Skills, err
}

type Standing struct {
	FromID   int     `json:"from_id"`
	FromType string  `json:"from_type"`
	Standing float64 `json:"standing"`
}

func (c *client) GetCharacterStandings(ctx context.Context, characterID int) (result []Standing, err error) {
	url := fmt.Sprintf("/characters/%d/standings/", characterID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}
