package util

import (
	"context"

	"github.com/nicklaw5/helix/v2"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

func GetNewSpotifyClient(ctx context.Context, a *AuthConfig, token *oauth2.Token) *spotify.Client {
	return spotify.New(a.OAuth.Client(ctx, token))
}

func RefreshSpotifyToken(ctx context.Context, a *AuthConfig, token *oauth2.Token) (*oauth2.Token, error) {
	source := a.OAuth.TokenSource(ctx, token)
	return source.Token()
}

func GetNewTwitchClient(a *AuthConfig) (*helix.Client, error) {
	opt := helix.Options{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		RedirectURI:  a.RedirectURL,
		UserAgent:    "TwitchSongRequests",
	}

	if a.APIBaseURL != "" {
		opt.APIBaseURL = a.APIBaseURL
	}

	return helix.NewClient(&opt)
}
