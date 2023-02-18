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

	clientSecret, err := GetFromEnv(constants.TwitchClientSecretKey)
	if err != nil {
		return nil, err
	}

	redirectURL, err := GetFromEnv(constants.TwitchRedirectURL)
	if err != nil {
		redirectURL = "localhost:8000"
	}

	opt := helix.Options{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURL,
		UserAgent:    "TwitchSongRequests",
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

	clientSecret, err := GetFromEnv(constants.SpotifyClientSecretKey)
	if err != nil {
		return nil, err
	}

	redirect, err := GetFromEnv(constants.SpotifyRedirectURL)
	if err != nil {
		return nil, err
	}

	return spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
		spotifyauth.WithRedirectURL(redirect)), nil
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

// GetFromEnvOrDefault tries to get the environment variable by the given key,
// and if the var is empty/undefined, it returns the supplied default
// value instead.
func GetFromEnvOrDefault(key, def string) string {
	s, err := GetFromEnv(key)
	if err != nil {
		return def
	}

	return s
}
