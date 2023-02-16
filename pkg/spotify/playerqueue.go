package spotify

import (
	"regexp"

	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
)

var openSpotifyURLPattern = regexp.MustCompile(`^https://open.spotify.com/track/([A-Za-z0-9]+)`)

// ensure struct implements queue.Publisher
var _ queue.Publisher = (*SpotifyPlayerQueue)(nil)

type SpotifyPlayerQueue struct {
}

// Publish will validate that the input matches a valid Spotify URL scheme,
// and then attempt to queue it in the user's Spotify player.
func (s *SpotifyPlayerQueue) Publish(client *spotify.Client, url string) error {
	return nil
}

// parseSpotifyTrackID takes an input string and tries to match it to the URL that you
// get from sharing a Spotify track externally
func parseSpotifyTrackID(s string) string {
	groups := openSpotifyURLPattern.FindStringSubmatch(s)
	if len(groups) < 1 {
		return ""
	}

	return groups[len(groups)-1]
}