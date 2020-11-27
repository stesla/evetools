package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stesla/evetools/sde"
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

func (c *client) GetStructure(ctx context.Context, id int) (result *sde.Station, err error) {
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
