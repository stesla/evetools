package esi

import (
	"context"
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
}

type client struct {
	http *http.Client
}

func NewClient(http *http.Client) Client {
	return &client{http: http}
}
