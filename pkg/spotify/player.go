package spotify

import (
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// ensure struct implements queue.Publisher
var _ queue.Publisher = (*SpotifyPlayerQueue)(nil)

type SpotifyPlayerQueue struct {
	// todo: need to create spotify client
	Auth *spotifyauth.Authenticator
}

func (s SpotifyPlayerQueue) Publish(val interface{}) error {
	return nil
}
