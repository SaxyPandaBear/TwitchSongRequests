package util

import (
	"fmt"
	"os"
	"strings"

	"github.com/saxypandabear/twitchsongrequests/internal/constants"
	"golang.org/x/oauth2"
)

const (
	// SpotifyAuthURL is the URL to Spotify Accounts Service's OAuth2 endpoint.
	SpotifyAuthURL = "https://accounts.spotify.com/authorize"

	//nolint:gosec // Mistakenly flagging an authorization URL
	// SpotifyTokenURL is the URL to the Spotify Accounts Service's OAuth2
	// token endpoint.
	SpotifyTokenURL = "https://accounts.spotify.com/api/token"
	// SpotifyUserScope is the set of permissions required to access the necessary
	// Spotify APIs
	SpotifyUserScope = "user-modify-playback-state user-read-playback-state user-read-email"
	// TwitchUserScope is the set of permissions required to access the necessary
	// Twitch APIs
	TwitchUserScope = "chat:read chat:edit channel:manage:redemptions"
)

// LoadTwitchConfigs reads from environment variables in order to
// populate configurations for creating a Twitch SDK client
func LoadTwitchConfigs() (*AuthConfig, error) {
	clientID, err := GetFromEnv(constants.TwitchClientIDKey)
	if err != nil {
		return nil, err
	}

	clientSecret, err := GetFromEnv(constants.TwitchClientSecretKey)
	if err != nil {
		return nil, err
	}

	state, err := GetFromEnv(constants.TwitchStateKey)
	if err != nil {
		return nil, err
	}

	// if deployed via railway, we want to derive the URL
	var redirectURL string

	railwayDomain, err := GetFromEnv(constants.RailwayDomain)
	if err == nil {
		redirectURL = fmt.Sprintf("https://%s/oauth/twitch", railwayDomain)
	} else {
		redirectURL = GetFromEnvOrDefault(constants.TwitchRedirectURL, "localhost:8000/oauth/twitch")
	}
	apiURL := GetFromEnvOrDefault(constants.MockServerURLKey, "")

	return &AuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		APIBaseURL:   apiURL,
		Scope:        TwitchUserScope,
		State:        state,
	}, nil
}

func LoadSpotifyConfigs() (*AuthConfig, error) {
	clientID, err := GetFromEnv(constants.SpotifyClientIDKey)
	if err != nil {
		return nil, err
	}

	clientSecret, err := GetFromEnv(constants.SpotifyClientSecretKey)
	if err != nil {
		return nil, err
	}

	// if deployed via railway, we want to derive the URL
	var redirect string

	railwayDomain, err := GetFromEnv(constants.RailwayDomain)
	if err == nil {
		redirect = fmt.Sprintf("https://%s/oauth/spotify", railwayDomain)
	} else {
		// TODO: idk if we want to fail here.
		var err2 error
		redirect, err2 = GetFromEnv(constants.SpotifyRedirectURL)
		if err2 != nil {
			return nil, err2
		}
	}

	state, err := GetFromEnv(constants.SpotifyStateKey)
	if err != nil {
		return nil, err
	}

	c := oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  SpotifyAuthURL,
			TokenURL: SpotifyTokenURL,
		},
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect,
		Scopes:       strings.Split(SpotifyUserScope, " "),
	}

	return &AuthConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        SpotifyUserScope,
		RedirectURL:  redirect,
		State:        state,
		OAuth:        &c,
	}, nil
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
