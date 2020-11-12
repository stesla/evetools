package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type HistoryDay struct {
	Date       string  `json:"date"`
	Lowest     float64 `json:"lowest"`
	Average    float64 `json:"average"`
	Highest    float64 `json:"highest"`
	OrderCount int64   `json:"order_count"`
	Volume     int64   `json:"volume"`
}

func (c *Client) MarketHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error) {
	url := fmt.Sprintf("/markets/%d/history/", regionID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)

	q := req.URL.Query()
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

type MarketOrder struct {
	Duration      int     `json:"duration"`
	Escrow        float64 `json:"escrow"`
	IsBuyOrder    bool    `json:"is_buy_order"`
	IsCorporation bool    `json:"is_corporation"`
	Issued        string  `json:"issued"`
	LocationID    int     `json:"location_id"`
	MinVolume     int     `json:"min_volume"`
	OrderID       int     `json:"order_id"`
	Price         float64 `json:"price"`
	Range         string  `json:"range"`
	RegionID      int     `json:"region_id"`
	TypeID        int     `json:"type_id"`
	VolumeRemain  int     `json:"volume_remain"`
	VolumeTotal   int     `json:"volume_total"`

	StationName   string `json:"station_name,omitempty"`
	TimeRemaining string `json:"time_remaining,omitempty"`
}

func (c *Client) MarketOrders(ctx context.Context, userID int) ([]*MarketOrder, error) {
	url := fmt.Sprintf("/characters/%d/orders/", userID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orders []*MarketOrder
	err = json.NewDecoder(resp.Body).Decode(&orders)
	return orders, err
}

func (c *Client) MarketPrices(ctx context.Context, stationID, regionID, typeID int) (*Price, error) {
	var buy, sell float64
	var page = 1

	for {
		url := fmt.Sprintf("/markets/%d/orders/", regionID)
		req, err := newESIRequest(ctx, http.MethodGet, url, nil)

		q := req.URL.Query()
		q.Add("order_type", "all")
		q.Add("type_id", strconv.Itoa(typeID))
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var orders []struct {
			IsBuyOrder bool    `json:"is_buy_order"`
			Price      float64 `json:"price"`
			LocationID int     `json:"location_id"`
		}
		err = json.NewDecoder(resp.Body).Decode(&orders)
		if err != nil {
			return nil, err
		}

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

		if str := resp.Header.Get("X-Pages"); str == "" {
			break
		} else if i, _ := strconv.Atoi(str); page >= i {
			break
		}
	}

	return &Price{Buy: buy, Sell: sell}, nil
}

func (c *Client) OpenMarketWindow(ctx context.Context, typeID int) (crr error) {
	req, err := newESIRequest(ctx, http.MethodPost, "/ui/openwindow/marketdetails/", nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("type_id", strconv.Itoa(typeID))
	req.URL.RawQuery = q.Encode()

	resp, err := c.http.Do(req)
	if err == nil && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
	resp.Body.Close()
	return
}