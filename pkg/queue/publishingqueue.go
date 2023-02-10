package queue

import "github.com/zmb3/spotify/v2"

type Publisher interface {
	// Publish takes a Spotify client and a value (the URL for the Spotify track)
	// and attempts to queue the song to the user's player. The client parameter is tied
	// to an individual user's access token.
	Publish(client *spotify.Client, url string) error
}
