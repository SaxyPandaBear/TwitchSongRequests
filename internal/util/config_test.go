package util_test

import (
	"strings"
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/util"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestGetFromEnvOrDefault(t *testing.T) {
	t.Setenv("foo", "bar")

	assert.Equal(t, "bar", util.GetFromEnvOrDefault("foo", "baz"))
	assert.Equal(t, "baz", util.GetFromEnvOrDefault("bar", "baz"))
}

func TestLoadTwitchConfigs(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "foo")
	t.Setenv(constants.TwitchClientSecretKey, "bar")
	t.Setenv(constants.TwitchStateKey, "baz")
	t.Setenv(constants.TwitchRedirectURL, "foo1")

	c, err := util.LoadTwitchConfigs()
	assert.NoError(t, err)
	assert.Equal(t, "foo", c.ClientID)
	assert.Equal(t, "bar", c.ClientSecret)
	assert.Equal(t, "baz", c.State)
	assert.Equal(t, util.TwitchUserScope, c.Scope)
	assert.Equal(t, "foo1", c.RedirectURL)
	assert.Empty(t, c.APIBaseURL)
	assert.Nil(t, c.OAuth)
}

func TestLoadTwitchConfigsWithMockAPI(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "foo")
	t.Setenv(constants.TwitchClientSecretKey, "bar")
	t.Setenv(constants.TwitchStateKey, "baz")
	t.Setenv(constants.TwitchRedirectURL, "foo1")
	t.Setenv(constants.MockServerURLKey, "bar2")

	c, err := util.LoadTwitchConfigs()
	assert.NoError(t, err)
	assert.Equal(t, "bar2", c.APIBaseURL)
}

func TestLoadTwitchConfigsWithDefaultRedirect(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "foo")
	t.Setenv(constants.TwitchClientSecretKey, "bar")
	t.Setenv(constants.TwitchStateKey, "baz")

	c, err := util.LoadTwitchConfigs()
	assert.NoError(t, err)
	assert.Equal(t, "localhost:8000", c.RedirectURL)
}

func TestLoadTwitchConfigsErrors(t *testing.T) {
	c, err := util.LoadTwitchConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.TwitchClientIDKey, "foo")
	c, err = util.LoadTwitchConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.TwitchClientSecretKey, "bar")
	c, err = util.LoadTwitchConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.TwitchStateKey, "baz")
	c, err = util.LoadTwitchConfigs()
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestLoadSpotifyConfigs(t *testing.T) {
	t.Setenv(constants.SpotifyClientIDKey, "foo")
	t.Setenv(constants.SpotifyClientSecretKey, "bar")
	t.Setenv(constants.SpotifyRedirectURL, "baz")
	t.Setenv(constants.SpotifyStateKey, "foo1")

	c, err := util.LoadSpotifyConfigs()
	assert.NoError(t, err)
	assert.NotNil(t, c)

	assert.Equal(t, "foo", c.ClientID)
	assert.Equal(t, "bar", c.ClientSecret)
	assert.Equal(t, "baz", c.RedirectURL)
	assert.Empty(t, c.APIBaseURL)
	assert.Equal(t, "foo1", c.State)
	assert.NotNil(t, c.OAuth)
	assert.Equal(t, util.SpotifyAuthURL, c.OAuth.Endpoint.AuthURL)
	assert.Equal(t, util.SpotifyTokenURL, c.OAuth.Endpoint.TokenURL)
	assert.Equal(t, "foo", c.OAuth.ClientID)
	assert.Equal(t, "bar", c.OAuth.ClientSecret)
	assert.Equal(t, "baz", c.OAuth.RedirectURL)
	assert.Equal(t, util.SpotifyUserScope, strings.Join(c.OAuth.Scopes, " "))
}

func TestLoadSpotifyConfigsErrors(t *testing.T) {
	c, err := util.LoadSpotifyConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.SpotifyClientIDKey, "foo")
	c, err = util.LoadSpotifyConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.SpotifyClientSecretKey, "bar")
	c, err = util.LoadSpotifyConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)
	t.Setenv(constants.SpotifyRedirectURL, "baz")
	c, err = util.LoadSpotifyConfigs()
	assert.Nil(t, c)
	assert.Error(t, err)

	t.Setenv(constants.SpotifyStateKey, "foo1")
	c, err = util.LoadSpotifyConfigs()
	assert.NotNil(t, c)
	assert.NoError(t, err)
}
