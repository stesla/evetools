package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/antihax/optional"
	"github.com/stesla/evetools/esi"
)

const (
	regionTheForge       = 10000002
	locationJitaTradeHub = 60003760
)

type ESIClient struct {
	api  *esi.APIClient
	http *http.Client
}

type contextToken int

const ESITokenKey contextToken = 1

func NewESIClient(client *http.Client) *ESIClient {
	cfg := esi.NewConfiguration()
	cfg.HTTPClient = client
	cfg.UserAgent = "evetools 0.0.1 - github.com/stesla/evetools - Stewart Cash"
	return &ESIClient{
		api:  esi.NewAPIClient(cfg),
		http: client,
	}
}

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	str := d.Format("2006-01-02")
	return json.Marshal(str)
}

type HistoryDay struct {
	Date       Date    `json:"date"`
	Lowest     float64 `json:"lowest"`
	Average    float64 `json:"average"`
	Highest    float64 `json:"highest"`
	OrderCount int64   `json:"order_count"`
	Volume     int64   `json:"volume"`
}

func (e *ESIClient) JitaHistory(ctx context.Context, typeID int) ([]HistoryDay, error) {
	days, _, err := e.api.MarketApi.GetMarketsRegionIdHistory(ctx, regionTheForge, int32(typeID), nil)
	if err != nil {
		return nil, err
	}

	result := make([]HistoryDay, len(days))
	for i, day := range days {
		hd := &result[i]
		t, _ := time.Parse("2006-01-02", day.Date)
		hd.Date = Date{t}
		hd.Lowest = day.Lowest
		hd.Average = day.Average
		hd.Highest = day.Highest
		hd.OrderCount = day.OrderCount
		hd.Volume = day.Volume
	}
	return result, nil
}

type Price struct {
	Buy, Sell float64
}

func (p Price) Margin() float64 {
	if p.Buy == 0 {
		return 0
	}
	return (p.Sell - p.Buy) / p.Buy * 100
}

func (e *ESIClient) JitaPrices(ctx context.Context, typeID int) (*Price, error) {
	opts := esi.MarketApiGetMarketsRegionIdOrdersOpts{
		TypeId: optional.NewInt32(int32(typeID)),
	}
	orders, _, err := e.api.MarketApi.GetMarketsRegionIdOrders(ctx, "all", regionTheForge, &opts)
	if err != nil {
		return nil, err
	}

	var buy, sell float64
	for _, order := range orders {
		if order.LocationId != locationJitaTradeHub {
			continue
		}
		if order.IsBuyOrder && order.Price > buy {
			buy = order.Price
		} else if !order.IsBuyOrder && (sell == 0 || order.Price < sell) {
			sell = order.Price
		}
	}
	return &Price{Buy: buy, Sell: sell}, nil
}

func (e *ESIClient) OpenMarketWindow(ctx context.Context, typeID int) (err error) {
	const apiURL = "https://esi.evetech.net/latest/ui/openwindow/marketdetails/"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, nil)
	if err != nil {
		return
	}

	q := url.Values{}
	q.Add("datasource", "tranquility")
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ctx.Value(ESITokenKey)))

	_, err = e.http.Do(req)
	return
}
