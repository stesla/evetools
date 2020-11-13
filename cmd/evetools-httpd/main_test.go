package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"

	"github.com/stesla/evetools/model"
	"github.com/stesla/evetools/sde"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(map[string]interface{}{})
	store = sessions.NewCookieStore([]byte{})
}

func TestMain(m *testing.M) {
	viper.Set("oauth.basePath", "https://esi")
	viper.Set("oauth.clientId", "CLIENT_ID")
	viper.Set("oauth.clientSecret", "CLIENT_SECRET")
	viper.Set("oauth.redirectURL", "REDIRECT_URL")
	initOAuthConfig()
	os.Exit(m.Run())
}

func TestCurrentUserUnauthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/user/current", nil)
	resp := handleRequest(t, req)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestLogin(t *testing.T) {
	assert := assert.New(t)
	req, _ := http.NewRequest("GET", "/login", nil)
	resp := handleRequest(t, req)

	assert.Equal(http.StatusFound, resp.StatusCode)

	loc, err := url.Parse(resp.Header.Get("Location"))
	if assert.NoError(err) {
		assert.Equal("https", loc.Scheme)
		assert.Equal("esi", loc.Host)
		assert.Equal("/v2/oauth/authorize", loc.Path)
		q := loc.Query()
		assert.Equal("code", q.Get("response_type"))
		assert.Equal("REDIRECT_URL", q.Get("redirect_uri"))
		assert.Equal("CLIENT_ID", q.Get("client_id"))
		assert.NotEmpty(q.Get("scope"), "scope")
		assert.NotEmpty(q.Get("state"), "state")
	}
}

func TestLoginCallback(t *testing.T) {
	assert := assert.New(t)

	mrt.AddHandler("https://esi/.well-known/oauth-authorization-server",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":   "https://esi",
				"jwks_uri": "https://esi/oauth/jwks",
			})
		}))

	privrsa, _ := rsa.GenerateKey(rand.Reader, 2048)
	privKey, _ := jwk.New(privrsa)
	pubKey, _ := jwk.New(privrsa.PublicKey)
	pubKey.Set(jwk.KeyUsageKey, string(jwk.ForSignature))
	pubKey.Set(jwk.KeyIDKey, "JWT-Signature-Key")
	mrt.AddHandler("https://esi/oauth/jwks",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jwk.Set{Keys: []jwk.Key{pubKey}})
		}))

	token := jwt.New()
	token.Set(jwt.SubjectKey, "CHARACTER:EVE:123456890")
	token.Set("name", "Bob Awox")
	token.Set("owner", "OWNER$TOKEN")
	headers := jws.NewHeaders()
	headers.Set(jwk.KeyIDKey, pubKey.KeyID())
	headers.Set(jwt.IssuerKey, "https://esi")
	compact, _ := jwt.Sign(token, jwa.RS256, privKey, jwt.WithHeaders(headers))
	mrt.AddHandler("https://esi/v2/oauth/token",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  string(compact),
				"expires_in":    1199,
				"token_type":    "Bearer",
				"refresh_token": "REFRESH_TOKEN",
			})
		}))

	req, _ := http.NewRequest("GET", "/login/callback", nil)
	q := req.URL.Query()
	q.Set("code", "CODE")
	q.Set("state", "STATE")
	req.URL.RawQuery = q.Encode()
	setSessionCookie(req, map[interface{}]interface{}{
		"oauth.state": "STATE",
	})
	resp := handleRequest(t, req)

	if assert.Equal(http.StatusFound, resp.StatusCode) {
		assert.Equal("/", resp.Header.Get("Location"))
		if cookies := resp.Cookies(); assert.Equal(1, len(cookies)) {
			name := viper.GetString("httpd.session.name")
			session := sessions.NewSession(store, name)
			c := cookies[0]
			err := securecookie.DecodeMulti(name, c.Value, &session.Values, store.Codecs...)
			if assert.NoError(err) {
				tok := session.Values["token"]
				if assert.IsType(tok, oauth2.Token{}) {
					assert.Equal(string(compact), session.Values["token"].(oauth2.Token).AccessToken)
				}
			}
		}
	}
}

func TestLogout(t *testing.T) {
	assert := assert.New(t)

	req, _ := http.NewRequest("GET", "/logout", nil)
	resp := handleRequest(t, req)

	if assert.Equal(http.StatusFound, resp.StatusCode) {
		assert.Equal("/", resp.Header.Get("Location"))
		if cookies := resp.Cookies(); assert.Equal(1, len(cookies)) {
			c := cookies[0]
			assert.Equal(-1, c.MaxAge)
			assert.True(c.Expires.Before(time.Now()))
		}
	}
}

func createSignedToken(claims map[string]interface{}) string {
	privRSA, _ := rsa.GenerateKey(rand.Reader, 2048)
	privKey, _ := jwk.New(privRSA)

	token := jwt.New()
	for key, value := range claims {
		token.Set(key, value)
	}
	compact, _ := jwt.Sign(token, jwa.RS256, privKey)
	return string(compact)
}

func TestCurrentUserAuthenticated(t *testing.T) {
	assert := assert.New(t)

	compact := createSignedToken(map[string]interface{}{
		jwt.SubjectKey: "CHARACTER:EVE:1234567890",
		"name":         "Bob Awox",
	})

	req, _ := http.NewRequest("GET", "/api/v1/user/current", nil)
	setSessionCookie(req, map[interface{}]interface{}{
		"token": &oauth2.Token{
			AccessToken:  string(compact),
			RefreshToken: "REFRESH",
		},
		"user": &model.User{
			ActiveCharacterID: 1234567890,
			StationID:         76543210,
		},
	})
	resp := handleRequest(t, req)

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var obj struct {
		Character   model.Character `json:"character"`
		StationName string          `json:"stationName"`
	}
	json.NewDecoder(resp.Body).Decode(&obj)
	assert.Equal(1234567890, obj.Character.ID)
	assert.Equal("Bob Awox", obj.Character.Name)
}

type failHandler struct {
	t *testing.T
}

func (h *failHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	h.t.Log("failHandler")
	h.t.FailNow()
}

func handleRequest(t *testing.T, r *http.Request) *http.Response {
	w := httptest.NewRecorder()
	s := NewServer(&failHandler{t}, &testDB{}, &testSDB{})
	s.http.Transport = mrt
	s.ServeHTTP(w, r)
	mrt.Reset()
	return w.Result()
}

type values map[interface{}]interface{}

func setSessionCookie(r *http.Request, vals values) {
	req := &http.Request{}
	session, _ := store.Get(req, viper.GetString("httpd.session.name"))
	for k, v := range vals {
		session.Values[k] = v
	}
	rec := httptest.NewRecorder()
	session.Save(req, rec)
	r.AddCookie(rec.Result().Cookies()[0])
}

type mockRoundTripper struct {
	handlers map[string]http.Handler
}

var mrt = &mockRoundTripper{handlers: make(map[string]http.Handler)}

func (m *mockRoundTripper) AddHandler(url string, handler http.Handler) {
	m.handlers[url] = handler
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	recorder := httptest.NewRecorder()
	if h := m.handlers[req.URL.String()]; h != nil {
		h.ServeHTTP(recorder, req)
	} else {
		http.NotFound(recorder, req)
	}
	return recorder.Result(), nil
}

func (m *mockRoundTripper) Reset() {
	for k := range m.handlers {
		delete(m.handlers, k)
	}
}

type testDB struct{}

var ErrNotImplemented = errors.New("not implemented")

func (*testDB) IsFavorite(int, int) (bool, error) {
	return false, ErrNotImplemented
}

func (*testDB) GetFavoriteTypes(int) ([]int, error) {
	return []int{587, 10244, 11198, 603}, nil
}

func (*testDB) SetFavorite(int, int, bool) error {
	return ErrNotImplemented
}

func (*testDB) FindOrCreateUserForCharacter(characterID int, characterName, owner string) (*model.User, error) {
	return &model.User{
		ID:                42,
		ActiveCharacterID: characterID,
	}, nil
	return nil, ErrNotImplemented
}

func (*testDB) GetCharacter(characterID int) (*model.Character, error) {
	return &model.Character{
		ID:   characterID,
		Name: "Bob Awox",
	}, nil
}

func (*testDB) SaveUserStation(userID, stationID int) error { return ErrNotImplemented }

type testSDB struct{}

func (*testSDB) GetMarketGroups() (map[int]*sde.MarketGroup, error)    { return nil, ErrNotImplemented }
func (*testSDB) GetMarketTypes() (map[int]*sde.MarketType, error)      { return nil, ErrNotImplemented }
func (*testSDB) GetStations(q string) (map[string]*sde.Station, error) { return nil, ErrNotImplemented }
func (*testSDB) GetStationByID(stationID int) (*sde.Station, error) {
	return &sde.Station{
		ID:     76543210,
		Region: sde.Region{ID: 12345678},
		System: sde.System{ID: 43218765},
		Name:   "Planet I - Moon 2 - Fake Station",
	}, nil
}
func (*testSDB) SearchTypesByName(filter string) ([]int, error) { return nil, ErrNotImplemented }
