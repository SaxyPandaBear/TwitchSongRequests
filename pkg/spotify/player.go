package spotify

import (
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type SpotifyPlayeQueue struct {
	// todo: need to create spotify client
	Auth *spotifyauth.Authenticator
}

func (s SpotifyPlayeQueue) Publish(val interface{}) error {
	return nil
}
