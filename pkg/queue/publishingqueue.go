package queue

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

// Queuer is the interface facade around the Spotify client to allow for local unit test mocking.
type Queuer interface {
	QueueSong(ctx context.Context, trackID spotify.ID) error
	GetTrack(ctx context.Context, id spotify.ID, opts ...spotify.RequestOption) (*spotify.FullTrack, error)
}

type Publisher interface {
	// Publish takes a Spotify client and a value (the URL for the Spotify track)
	// and attempts to queue the song to the user's player. The client parameter is tied
	// to an individual user's access token.
	Publish(client Queuer, url string, allowExplicit bool) error
}
