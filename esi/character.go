package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

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
