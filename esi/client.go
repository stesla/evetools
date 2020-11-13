package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client interface {
	GetCharacterName(characterID int) (string, error)
	GetMarketOrderHistory(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetMarketOrders(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetMarketPrices(ctx context.Context, stationID, regionID, typeID int) (*Price, error)
	GetPriceHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error)
	GetWalletBalance(ctx context.Context, characterID int) (balance float64, err error)
	GetWalletTransactions(ctx context.Context, characterID int) ([]*WalletTransaction, error)
	OpenMarketWindow(ctx context.Context, typeID int) (crr error)
}

type client struct {
	http *http.Client
}

func NewClient(http *http.Client) Client {
	return &client{http: http}
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
