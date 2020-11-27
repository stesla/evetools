package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/stesla/evetools/config"
)

type contextKey int

const AccessTokenKey contextKey = 1

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	str := d.Format("2006-01-02")
	return json.Marshal(str)
}

func newESIRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	return newESIRequestWithURL(ctx, method, config.EsiBasePath()+"/latest"+path, body)

}

func newMetaRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	return newESIRequestWithURL(ctx, method, config.EsiBasePath()+path, body)
}

func newESIRequestWithURL(ctx context.Context, method, addr string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("datasource", "tranquility")
	req.URL.RawQuery = q.Encode()

	token := ctx.Value(AccessTokenKey)
	if token != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	// TODO: maybe have a version variable?
	req.Header.Add("User-Agent", "evetools 0.0.1 - github.com/stesla/evetools - Stewart Cash")
	return req, nil
}
