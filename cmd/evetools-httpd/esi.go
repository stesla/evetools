package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strconv"
	"time"
)

type ESIClient struct {
	http *http.Client
}

func NewESIClient(client *http.Client) *ESIClient {
	return &ESIClient{
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
	Date       string  `json:"date"`
	Lowest     float64 `json:"lowest"`
	Average    float64 `json:"average"`
	Highest    float64 `json:"highest"`
	OrderCount int64   `json:"order_count"`
	Volume     int64   `json:"volume"`
}

func (e *ESIClient) MarketHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error) {
	url := fmt.Sprintf("/markets/%d/history/", regionID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)

	q := req.URL.Query()
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return
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

type marketOrder struct {
	Duration     int     `json:"duration"`
	IsBuyOrder   bool    `json:"is_buy_order"`
	Issued       string  `json:"issued"`
	LocationID   int     `json:"location_id"`
	MinVolume    int     `json:"min_volume"`
	OrderID      int     `json:"order_id"`
	Price        float64 `json:"price"`
	Range        string  `json:"range"`
	SystemID     int     `json:"system_id"`
	TypeID       int     `json:"type_id"`
	VolumeRemain int     `json:"volume_remain"`
	VolumeTotal  int     `json:"volume_total"`
}

func (e *ESIClient) MarketPrices(ctx context.Context, stationID, regionID, typeID int) (*Price, error) {
	// TODO: this could potentially have a paginated response
	url := fmt.Sprintf("/markets/%d/orders/", regionID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)

	q := req.URL.Query()
	q.Add("order_type", "all")
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	resp, err := e.http.Do(req)
	if err != nil {
		return nil, err
	}

	var orders []marketOrder
	err = json.NewDecoder(resp.Body).Decode(&orders)
	if err != nil {
		return nil, err
	}

	var buy, sell float64
	for _, order := range orders {
		if order.LocationID != stationID {
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
	req, err := newESIRequest(ctx, http.MethodPost, "/ui/openwindow/marketdetails/", nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("datasource", "tranquility")
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	_, err = e.http.Do(req)
	return
}

func newESIRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := "https://esi.evetech.net/latest" + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	q := neturl.Values{}
	q.Add("datasource", "tranquility")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", ctx.Value(ESITokenKey)))
	// TODO: maybe have a version variable?
	req.Header.Add("User-Agent", "evetools 0.0.1 - github.com/stesla/evetools - Stewart Cash")
	return req, nil

}
