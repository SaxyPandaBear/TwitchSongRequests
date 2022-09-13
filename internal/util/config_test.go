package util_test

import (
	"os"
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
	os.Setenv(constants.TwitchClientIDKey, "")
	defer os.Unsetenv(constants.TwitchClientIDKey)

	opts, err := util.LoadTwitchClientOptions()
	assert.Nil(t, opts)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty")
}

func TestMockServerUrlMissing(t *testing.T) {
	os.Setenv(constants.TwitchClientIDKey, "foo")
	defer os.Unsetenv(constants.TwitchClientIDKey)

	opts, err := util.LoadTwitchClientOptions()
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	assert.Equal(t, "foo", opts.ClientID)
	assert.Empty(t, opts.APIBaseURL)
}

func TestMockServerUrlPresent(t *testing.T) {
	os.Setenv(constants.TwitchClientIDKey, "foo")
	os.Setenv(constants.MockServerURLKey, "bar")
	defer os.Unsetenv(constants.TwitchClientIDKey)
	defer os.Unsetenv(constants.MockServerURLKey)

	opts, err := util.LoadTwitchClientOptions()
	assert.NotNil(t, opts)
	assert.NoError(t, err)
	assert.Equal(t, "foo", opts.ClientID)
	assert.Equal(t, "bar", opts.APIBaseURL)
}

func TestSpotifyClientMissing(t *testing.T) {
	auth, err := util.LoadSpotifyClientOptions()
	assert.Nil(t, auth)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "not defined")
}

func TestSpotifyClientEmpty(t *testing.T) {
	os.Setenv(constants.SpotifyClientIDKey, "")
	defer os.Unsetenv(constants.SpotifyClientIDKey)

	auth, err := util.LoadSpotifyClientOptions()
	assert.Nil(t, auth)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "empty")
}

func TestSpotifyClientHappyPath(t *testing.T) {
	os.Setenv(constants.SpotifyClientIDKey, "foo")
	defer os.Unsetenv(constants.SpotifyClientIDKey)

	auth, err := util.LoadSpotifyClientOptions()
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}
