package queue

import (
	"context"

	"github.com/zmb3/spotify/v2"
)

type Queuer interface {
	QueueSong(ctx context.Context, trackID spotify.ID) error
}

type Publisher interface {
	// Publish takes a Spotify client and a value (the URL for the Spotify track)
	// and attempts to queue the song to the user's player. The client parameter is tied
	// to an individual user's access token.
	Publish(client Queuer, url string) error
}
