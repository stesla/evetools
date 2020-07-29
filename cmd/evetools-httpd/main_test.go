package main

import (
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type failHandler struct {
	t *testing.T
}

func init() {
	gob.Register(map[string]interface{}{})
	store = sessions.NewCookieStore([]byte{})
}

func (h *failHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	h.t.Log("failHandler")
	h.t.FailNow()
}

func TestCurrentUserUnauthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/api/v1/currentUser", nil)
	rec := httptest.NewRecorder()
	h := makeHandler(&failHandler{t})
	h.ServeHTTP(rec, req)
	resp := rec.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	var obj map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&obj)
	assert.Equal(t, map[string]interface{}{}, obj)
}

func TestCurrentUserAuthenticated(t *testing.T) {
	req := &http.Request{}
	session, err := store.Get(req, viper.GetString("httpd.session.name"))
	assert.NoError(t, err)
	expected := map[string]interface{}{
		"foo": "bar",
		"baz": "quux",
	}
	session.Values["currentUser"] = expected
	rec := httptest.NewRecorder()
	err = session.Save(req, rec)
	assert.NoError(t, err)
	resp := rec.Result()
	cookies := resp.Cookies()
	assert.Equal(t, 1, len(cookies))

	req, _ = http.NewRequest("GET", "http://example.com/api/v1/currentUser", nil)
	req.AddCookie(cookies[0])
	rec = httptest.NewRecorder()
	h := makeHandler(&failHandler{t})
	h.ServeHTTP(rec, req)
	resp = rec.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	var obj map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&obj)
	assert.Equal(t, expected, obj)
}
