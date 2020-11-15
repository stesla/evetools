package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *client) GetWalletBalance(ctx context.Context, characterID int) (balance float64, err error) {
	url := fmt.Sprintf("/characters/%d/wallet/", characterID)
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

type WalletTransaction struct {
	ClientID      int     `json:"client_id"`
	ClientName    string  `json:"client_name,omitempty"`
	Date          string  `json:"date"`
	IsBuy         bool    `json:"is_buy"`
	IsPersonal    bool    `json:"is_personal"`
	JournalRefID  int     `json:"journal_ref_id"`
	LocationID    int     `json:"location_id"`
	Quantity      int     `json:"quantity"`
	TransactionID int     `json:"transaction_id"`
	TypeID        int     `json:"type_id"`
	UnitPrice     float64 `json:"unit_price"`
}

func (c *client) GetWalletTransactions(ctx context.Context, characterID int) (result []*WalletTransaction, err error) {
	url := fmt.Sprintf("/characters/%d/wallet/transactions/", characterID)
	req, err := newESIRequest(ctx, http.MethodGet, url, nil)
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
