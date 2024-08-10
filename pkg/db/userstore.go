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

type NoopUserStore struct{}

// AddUser implements UserStore.
func (n *NoopUserStore) AddUser(user *users.User) error {
	return nil
}

// DeleteUser implements UserStore.
func (n *NoopUserStore) DeleteUser(id string) error {
	return nil
}

// GetUser implements UserStore.
func (n *NoopUserStore) GetUser(id string) (*users.User, error) {
	return nil, nil
}

// UpdateUser implements UserStore.
func (n *NoopUserStore) UpdateUser(user *users.User) error {
	return nil
}

var _ UserStore = (*NoopUserStore)(nil)

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
