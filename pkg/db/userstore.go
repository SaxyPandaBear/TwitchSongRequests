package db

import (
	"github.com/saxypandabear/twitchsongrequests/pkg/users"
	"golang.org/x/oauth2"
)

type UserStore interface {
	GetUser(id string) (*users.User, error)
	AddUser(user *users.User) error
	UpdateUser(user *users.User) error
	DeleteUser(id string) error
}

func FetchSpotifyToken(userStore UserStore, id string) (*oauth2.Token, error) {
	u, err := userStore.GetUser(id)
	if err != nil {
		return nil, err
	}

	tok := oauth2.Token{
		AccessToken:  u.SpotifyAccessToken,
		RefreshToken: u.SpotifyRefreshToken,
		Expiry:       *u.SpotifyExpiry,
	}

	return &tok, nil
}

func FetchTwitchToken(userStore UserStore, id string) (*oauth2.Token, error) {
	u, err := userStore.GetUser(id)
	if err != nil {
		return nil, err
	}

	tok := oauth2.Token{
		AccessToken:  u.TwitchAccessToken,
		RefreshToken: u.TwitchRefreshToken,
	}

	return &tok, nil
}
