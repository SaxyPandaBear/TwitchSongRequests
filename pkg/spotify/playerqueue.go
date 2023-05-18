package spotify

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
)

var (
	openSpotifyURLPattern = regexp.MustCompile(`^https://open.spotify.com/(.+?/)*?track/([A-Za-z0-9]+)`)
	ErrInvalidInput       = errors.New("invalid user input for Spotify URI")
	ErrExplicitSong       = errors.New("user does not allow adding explicit songs to the queue")
	ErrSongTooLong        = errors.New("song is too long")
)

// ensure struct implements queue.Publisher
var _ queue.Publisher = (*SpotifyPlayerQueue)(nil)

// TODO: Publisher is an unnecessary struct because there is no state that the publisher tracks here.
type SpotifyPlayerQueue struct {
}

// Publish will validate that the input matches a valid Spotify URL scheme,
// and then attempt to queue it in the user's Spotify player.
func (s *SpotifyPlayerQueue) Publish(client queue.Queuer, url string, pref *preferences.Preference) error {
	id := parseSpotifyTrackID(url)
	if len(id) < 1 {
		return ErrInvalidInput
	}

	sID := spotify.ID(id)

	if err := ShouldQueue(client, sID, pref); err != nil {
		return err
	}

	return client.QueueSong(context.Background(), sID)
}

// TODO: this should be in the queuer
func ShouldQueue(client queue.Queuer, id spotify.ID, p *preferences.Preference) error {
	track, err := client.GetTrack(context.Background(), id)
	if err != nil {
		return fmt.Errorf("failed to get track %s: %w", id.String(), err)
	}

	if (p == nil || !p.ExplicitSongs) && track.Explicit {
		return ErrExplicitSong
	}

	if p != nil && p.MaxSongLength > 0 && track.Duration > p.MaxSongLength {
		return fmt.Errorf("song is too long. %d > %d", track.Duration, p.MaxSongLength)
	}

	return nil
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
