package util_test

import (
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/util"

	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestTwitchClientMissing(t *testing.T) {
	opts, err := util.LoadTwitchClientOptions()
	assert.Nil(t, opts)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not defined")
}

func TestTwitchClientEmpty(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "")

	opts, err := util.LoadTwitchClientOptions()
	assert.Nil(t, opts)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty")
}

func TestMockServerUrlMissing(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "foo")
	t.Setenv(constants.TwitchClientSecretKey, "bar")

	opts, err := util.LoadTwitchClientOptions()
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	assert.Equal(t, "foo", opts.ClientID)
	assert.Empty(t, opts.APIBaseURL)
}

func TestMockServerUrlPresent(t *testing.T) {
	t.Setenv(constants.TwitchClientIDKey, "foo")
	t.Setenv(constants.TwitchClientSecretKey, "bar")
	t.Setenv(constants.MockServerURLKey, "baz")

	opts, err := util.LoadTwitchClientOptions()
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	assert.Equal(t, "foo", opts.ClientID)
	assert.Equal(t, "bar", opts.ClientSecret)
	assert.Equal(t, "baz", opts.APIBaseURL)
}

func TestSpotifyClientMissing(t *testing.T) {
	auth, err := util.LoadSpotifyClientOptions()
	assert.Nil(t, auth)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not defined")
}

func TestSpotifyClientEmpty(t *testing.T) {
	t.Setenv(constants.SpotifyClientIDKey, "")

	auth, err := util.LoadSpotifyClientOptions()
	assert.Nil(t, auth)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty")
}

func TestSpotifyClientHappyPath(t *testing.T) {
	t.Setenv(constants.SpotifyClientIDKey, "foo")
	t.Setenv(constants.SpotifyClientSecretKey, "bar")

	auth, err := util.LoadSpotifyClientOptions()
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}

func TestGetFromEnvOrDefault(t *testing.T) {
	t.Setenv("foo", "bar")

	assert.Equal(t, "bar", util.GetFromEnvOrDefault("foo", "baz"))
	assert.Equal(t, "baz", util.GetFromEnvOrDefault("bar", "baz"))
}
