package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/stesla/evetools/sde"
)

type Client interface {
	GetCharacterName(characterID int) (string, error)
	GetCharacterSkills(ctx context.Context, characterID int) ([]Skill, error)
	GetCharacterStandings(ctx context.Context, characterID int) ([]Standing, error)
	GetCharacterOrderHistory(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetCharacterOrders(ctx context.Context, characterID int) ([]*MarketOrder, error)
	GetMarketPriceForType(stationID, regionID, typeID int) (*Price, error)
	GetNames(ids []int) (map[int]string, error)
	GetPriceHistory(ctx context.Context, regionID, typeID int) (result []HistoryDay, err error)
	GetStructure(ctx context.Context, id int) (*sde.Station, error)
	GetStructures() ([]int, error)
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

func (c *client) GetNames(ids []int) (map[int]string, error) {
	idMap := make(map[int]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}
	ids = []int{}
	for id, _ := range idMap {
		ids = append(ids, id)
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(&ids)

	req, err := newESIRequest(context.Background(), http.MethodPost, "/universe/names/", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	results := []struct {
		Category string `json:"category"`
		ID       int    `json:"id"`
		Name     string `json:"name"`
	}{}
	json.NewDecoder(resp.Body).Decode(&results)

	names := make(map[int]string, len(results))
	for _, r := range results {
		if r.Category != "character" {
			continue
		}
		names[r.ID] = r.Name
	}

	return names, nil
}
