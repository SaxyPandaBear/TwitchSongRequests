package api_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/saxypandabear/twitchsongrequests/pkg/api"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/saxypandabear/twitchsongrequests/pkg/db"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"github.com/stretchr/testify/assert"
)

func TestRevokeUserAccess(t *testing.T) {
	u := db.InMemoryUserStore{
		Data: map[string]*users.User{
			"123": {
				TwitchID: "123",
			},
		},
	}

	h := api.NewUserHandler(&u, "")

	req := httptest.NewRequest("DELETE", "/", nil)

	// TODO: until setup-go GitHub action supports Go 1.20,
	//       need to include a cookie expiry value.
	req.AddCookie(&http.Cookie{
		Name:    constants.TwitchIDCookieKey,
		Value:   base64.StdEncoding.EncodeToString([]byte("123")),
		Expires: time.Date(2099, time.April, 1, 2, 3, 4, 5, time.UTC),
	})

	rr := httptest.NewRecorder()
	api := http.HandlerFunc(h.RevokeUserAccesses)

	c := make(chan struct{})
	go func() {
		api.ServeHTTP(rr, req)
		c <- struct{}{}
	}()

	select {
	case <-c:
		t.Log("finished request")
	case <-time.After(25 * time.Millisecond):
		t.Error("did not receive message in time")
	}

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Empty(t, u.Data)
}
