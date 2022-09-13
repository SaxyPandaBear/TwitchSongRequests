package util

import (
	"fmt"
	"log"
	"os"

	"github.com/nicklaw5/helix"
	"github.com/saxypandabear/twitchsongrequests/pkg/constants"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// LoadTwitchClientOptions reads from environment variables in order to populate
// configuration options for the Twitch client.
func LoadTwitchClientOptions() (*helix.Options, error) {
	clientID, err := GetFromEnv(constants.TwitchClientIDKey)
	if err != nil {
		return nil, err
	}

	opt := helix.Options{
		ClientID: clientID,
	}

	// handle case where we are running locally with the Twitch CLI mock server
	url, err := GetFromEnv(constants.MockServerURLKey)
	if err == nil {
		log.Printf("Using mocked Twitch API hosted at %s\n", url)
		opt.APIBaseURL = url
	}

	return &opt, nil
}

// LoadSpotifyClientOptions reads from environment variables in order to populate
// configuration options for the Spotify authenticator
func LoadSpotifyClientOptions() (*spotifyauth.Authenticator, error) {
	clientID, err := GetFromEnv(constants.SpotifyClientIDKey)
	if err != nil {
		return nil, err
	}

	return spotifyauth.New(spotifyauth.WithClientID(clientID)), nil
}

// GetFromEnv tries to read an environment variable by key, and if:
// 1. the key does not exist: return an error
// 2. the key exists but the value is empty: return an error
// 3. the key exists AND the value is non-empty: return the value
func GetFromEnv(key string) (string, error) {
	s, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%s is not defined in the environment", key)
	} else if len(s) < 1 {
		return "", fmt.Errorf("%s is defined, but empty", key)
	}

	return s, nil
}
