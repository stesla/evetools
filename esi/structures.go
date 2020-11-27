package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *client) GetStructures() (result []int, err error) {
	ctx := context.Background()
	req, err := newESIRequest(ctx, http.MethodGet, "/universe/structures/", nil)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}

type Structure struct {
	ID       int    `josn:"id"`
	Name     string `json:"name"`
	CorpID   int    `json:"owner_id"`
	SystemID int    `json:"solar_system_id"`
	RegionID int    `json:"region_id"`
}

func (c *client) GetStructure(ctx context.Context, id int) (result *Structure, err error) {
	url := fmt.Sprintf("/universe/structures/%d/", id)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}
