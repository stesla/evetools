package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"

	"github.com/stesla/evetools/config"
	"github.com/stesla/evetools/esi"
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
	viper.Set("esi.basePath", "https://esi")
	viper.Set("oauth.basePath", "https://esi")
	viper.Set("oauth.clientId", "CLIENT_ID")
	viper.Set("oauth.clientSecret", "CLIENT_SECRET")
	viper.Set("oauth.redirectURL", "REDIRECT_URL")
	viper.Set("cli.oauth.clientId", "CLIENT_ID")
	viper.Set("cli.oauth.clientSecret", "CLIENT_SECRET")
	viper.Set("cli.oauth.redirectURL", "REDIRECT_URL")
	config.InitializeOAuth()
	os.Exit(m.Run())
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
		assert.NotEmpty(q.Get("state"), "state")
	}
}

func TestLoginCallback(t *testing.T) {
	assert := assert.New(t)

	// stub out call from oauth2.Exchange
	privrsa, _ := rsa.GenerateKey(rand.Reader, 2048)
	privKey, _ := jwk.New(privrsa)
	pubKey, _ := jwk.New(privrsa.PublicKey)
	pubKey.Set(jwk.KeyUsageKey, string(jwk.ForSignature))
	pubKey.Set(jwk.KeyIDKey, "JWT-Signature-Key")
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
				"refresh_token": "REFRESH",
			})
		}))

	// stub out call from esi.Verfiy
	mrt.AddHandler("https://esi/verify/?datasource=tranquility",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "applicaton/json")
			json.NewEncoder(w).Encode(esi.VerifyOK{})
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
				if assert.IsType(oauth2.Token{}, tok) {
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

type failHandler struct {
	t *testing.T
}

func (h *failHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	h.t.Log("failHandler")
	h.t.FailNow()
}

func handleRequest(t *testing.T, r *http.Request) *http.Response {
	w := httptest.NewRecorder()
	s := NewServer(&failHandler{t}, &testDB{}, &testViewRenderer{})
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
		http.Error(recorder, fmt.Sprintf("not found: %q", req.URL.String()), http.StatusNotFound)
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

func (*testDB) AllUserStations() ([]*sde.Station, error) { return nil, ErrNotImplemented }
func (*testDB) DeleteFavorites(userID int) (err error)   { return ErrNotImplemented }

func (*testDB) GetFavoriteTypes(int) ([]int, error) {
	return []int{587, 10244, 11198, 603}, nil
}

func (*testDB) IsFavorite(int, int) (bool, error) { return false, ErrNotImplemented }
func (*testDB) SetFavorite(int, int, bool) error  { return ErrNotImplemented }

func (*testDB) FindOrCreateCharacterForUser(int, esi.VerifyOK) (*model.Character, error) {
	return nil, ErrNotImplemented
}

func (db *testDB) FindOrCreateUserAndCharacter(verify esi.VerifyOK) (*model.User, *model.Character, error) {
	character, _ := db.GetCharacterByOwnerHash(verify.CharacterOwnerHash)
	user, _ := db.GetUser(character.UserID)
	return user, character, nil
}

func (*testDB) GetCharacterByUserAndCharacterID(int, int) (*model.Character, error) {
	return nil, ErrNotImplemented
}

func (*testDB) GetCharacterByOwnerHash(hash string) (*model.Character, error) {
	return &model.Character{
		ID:                 1,
		CharacterID:        1234567890,
		CharacterName:      "Bob Awox",
		CharacterOwnerHash: hash,
		UserID:             42,
	}, nil
}

func (*testDB) GetCharactersForUser(userID int) (map[int]*model.Character, error) {
	return nil, ErrNotImplemented
}

func (*testDB) GetLatestTxnID() (int, error) { return 0, ErrNotImplemented }

func (*testDB) GetPricesForStation(int) (map[int]esi.Price, error) {
	return nil, ErrNotImplemented
}

func (*testDB) GetTokenForCharacter(characterID int) (*model.Token, error) {
	return &model.Token{
		ID:           13,
		CharacterID:  characterID,
		RefreshToken: "REFRESH",
		Scopes:       strings.Join(config.OAuthForHTTP.Scopes, " "),
	}, nil
}

func (m *testDB) GetUser(userID int) (*model.User, error) {
	return &model.User{
		ID:                  userID,
		ActiveCharacterHash: "OWNER-HASH",
		ActiveCharacterID:   1234567890,
	}, nil
}

func (*testDB) RemoveCharacterForUser(int, int) error           { return ErrNotImplemented }
func (*testDB) SaveActiveCharacterHash(int, string) error       { return ErrNotImplemented }
func (*testDB) SavePrice(int, int, esi.Price) error             { return ErrNotImplemented }
func (*testDB) SaveTokenForCharacter(int, string, string) error { return nil }
func (*testDB) SaveUserStationA(userID, stationID int) error    { return ErrNotImplemented }
func (*testDB) SaveUserStationB(userID, stationID int) error    { return ErrNotImplemented }

type testViewRenderer struct{}

func (*testViewRenderer) renderView(http.ResponseWriter, *http.Request, string, template.FuncMap, interface{}) {
}
