package spotify

import (
	"context"
	"errors"
	"log"
	"regexp"

	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
)

var (
	openSpotifyURLPattern = regexp.MustCompile(`^https://open.spotify.com/track/([A-Za-z0-9]+)`)
	ErrInvalidInput       = errors.New("invalid user input for Spotify URI")
)

// ensure struct implements queue.Publisher
var _ queue.Publisher = (*SpotifyPlayerQueue)(nil)

// TODO: Publisher is an unnecessary struct because there is no state that the publisher tracks here.
type SpotifyPlayerQueue struct {
}

// Publish will validate that the input matches a valid Spotify URL scheme,
// and then attempt to queue it in the user's Spotify player.
func (s *SpotifyPlayerQueue) Publish(client queue.Queuer, url string, allowExplicit bool) error {
	id := parseSpotifyTrackID(url)
	if len(id) < 1 {
		return ErrInvalidInput
	}

	return client.QueueSong(context.Background(), spotify.ID(id))
}

// TODO: this should be in the queuer
func ShouldQueue(client queue.Queuer, id spotify.ID, allowExplicit bool) bool {
	track, err := client.GetTrack(context.Background(), id)
	if err != nil {
		log.Println("failed to get track", id.String(), err)
		return false
	}

	if !allowExplicit && track.Explicit {
		return false
	}

	return true
}

// parseSpotifyTrackID takes an input string and tries to match it to the URL that you
// get from sharing a Spotify track externally
// TODO: make this implemented by the queuer
func parseSpotifyTrackID(s string) string {
	groups := openSpotifyURLPattern.FindStringSubmatch(s)
	if len(groups) < 1 {
		return ""
	}

	return groups[len(groups)-1]
}
