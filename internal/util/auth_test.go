package util_test

import (
	"fmt"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/saxypandabear/twitchsongrequests/pkg/testutil"
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"github.com/stretchr/testify/assert"
)

func TestIsUserAuthenticated(t *testing.T) {
	u := testutil.InMemoryUserStore{
		Data: map[string]*users.User{
			"123": {
				TwitchID:            "123",
				TwitchAccessToken:   "foo",
				TwitchRefreshToken:  "foo",
				SpotifyAccessToken:  "bar",
				SpotifyRefreshToken: "bar",
			},
			"234": { // missing spotify auth
				TwitchID:           "234",
				TwitchAccessToken:  "foo",
				TwitchRefreshToken: "foo",
			},
		},
	}

	assert.True(t, util.IsUserAuthenticated(&u, "123"))
	assert.False(t, util.IsUserAuthenticated(&u, "234"))
	assert.False(t, util.IsUserAuthenticated(&u, "345"))
}

func TestGenerateAuthURL(t *testing.T) {
	c := util.AuthConfig{
		ClientID:    "foo",
		RedirectURL: "bar",
		State:       "baz",
		Scope:       "scope",
	}

	expected := fmt.Sprintf("https://some-host.com/path?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		c.ClientID, c.RedirectURL, c.Scope, c.State)

	actual := util.GenerateAuthURL("some-host.com", "/path", &c)
	assert.Equal(t, expected, actual)
}
