package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type contextKey int

const AccessTokenKey contextKey = 1

type Client struct {
	http *http.Client
}

func NewClient(client *http.Client) *Client {
	return &Client{
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

func newESIRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	addr := "https://esi.evetech.net/latest" + path
	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("datasource", "tranquility")
	req.URL.RawQuery = q.Encode()

	token := ctx.Value(AccessTokenKey)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// TODO: maybe have a version variable?
	req.Header.Add("User-Agent", "evetools 0.0.1 - github.com/stesla/evetools - Stewart Cash")
	return req, nil
}
