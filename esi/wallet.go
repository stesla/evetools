package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) WalletBalance(ctx context.Context, userID int) (balance float64, err error) {
	url := fmt.Sprintf("/characters/%d/wallet/", userID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&balance)
	return
}
