package esi

import (
	"context"
	"encoding/json"
	"net/http"
)

type Client interface {
	GetCharacterName(characterID int) (string, error)
	GetCharacterSkills(ctx context.Context, characterID int) ([]Skill, error)
	GetCharacterStandings(ctx context.Context, characterID int) ([]Standing, error)
	GetMarketOrderHistory(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetMarketOrders(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetMarketPrices(ctx context.Context, stationID, regionID, typeID int) (*Price, error)
	GetPriceHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error)
	GetWalletBalance(ctx context.Context, characterID int) (balance float64, err error)
	GetWalletTransactions(ctx context.Context, characterID int) ([]*WalletTransaction, error)
	OpenMarketWindow(ctx context.Context, typeID int) (crr error)
	Verify(ctx context.Context) (VerifyOK, error)
}

type client struct {
	http *http.Client
}

func NewClient(http *http.Client) Client {
	return &client{http: http}
}

type VerifyOK struct {
	CharacterID        int
	CharacterName      string
	CharacterOwnerHash string
	ClientID           string
	ExpiresOn          string
	Scopes             string
	TokenType          string
}

func (c *client) Verify(ctx context.Context) (result VerifyOK, err error) {
	req, err := newMetaRequest(ctx, http.MethodGet, "/verify/", nil)
	if err != nil {
		return
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	return
}
