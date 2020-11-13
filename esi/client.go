package esi

import (
	"context"
	"net/http"
)

type Client interface {
	GetWalletBalance(ctx context.Context, userID int) (balance float64, err error)
	GetPriceHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error)
	GetMarketOrders(ctx context.Context, userID int) ([]*MarketOrder, error)
	GetMarketOrderHistory(ctx context.Context, userID int) ([]*MarketOrder, error)
	GetMarketPrices(ctx context.Context, stationID, regionID, typeID int) (*Price, error)
	OpenMarketWindow(ctx context.Context, typeID int) (crr error)
}

type client struct {
	http *http.Client
}

func NewClient(http *http.Client) Client {
	return &client{http: http}
}
