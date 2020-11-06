package main

import (
	"context"
	"net/http"

	"github.com/antihax/optional"
	"github.com/stesla/evetools/esi"
)

const (
	regionTheForge       = 10000002
	locationJitaTradeHub = 60003760
)

type ESIClient struct {
	api *esi.APIClient
}

func NewESIClient(client *http.Client) *ESIClient {
	cfg := esi.NewConfiguration()
	cfg.HTTPClient = client
	cfg.UserAgent = "evetools 0.0.1 - github.com/stesla/evetools - Stewart Cash"
	return &ESIClient{
		api: esi.NewAPIClient(cfg),
	}
}

func (e *ESIClient) JitaHistory(ctx context.Context, typeID int) (volume, lowest, average, highest int, err error) {
	days, _, err := e.api.MarketApi.GetMarketsRegionIdHistory(ctx, regionTheForge, int32(typeID), nil)
	if err != nil {
		return
	}

	for _, day := range days {
		volume += int(day.Volume)
		lowest += int(day.Lowest)
		average += int(day.Average)
		highest += int(day.Highest)
	}
	volume /= len(days)
	lowest /= len(days)
	average /= len(days)
	highest /= len(days)
	return
}

func (e *ESIClient) JitaPrices(ctx context.Context, typeID int) (buy, sell float64, err error) {
	opts := esi.MarketApiGetMarketsRegionIdOrdersOpts{
		TypeId: optional.NewInt32(int32(typeID)),
	}
	orders, _, err := e.api.MarketApi.GetMarketsRegionIdOrders(ctx, "all", regionTheForge, &opts)
	if err != nil {
		return 0, 0, err
	}

	for _, order := range orders {
		if order.IsBuyOrder && order.Price > buy {
			buy = order.Price
		} else if !order.IsBuyOrder && (sell == 0 || order.Price < sell) {
			sell = order.Price
		}
	}
	return
}
