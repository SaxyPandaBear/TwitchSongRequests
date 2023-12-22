package spotify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/saxypandabear/twitchsongrequests/pkg/queue"
	"github.com/zmb3/spotify/v2"
)

var (
	openSpotifyURLPattern = regexp.MustCompile(`https://open.spotify.com/(.+?/)*?track/([A-Za-z0-9]+)`)
	// New Spotify URLs don't include the track ID in them anymore because they hate developers
	// https://community.spotify.com/t5/Spotify-for-Developers/How-to-see-through-shortened-https-spotify-link-links/m-p/5521244
	opaqueSpotifyURLPattern = regexp.MustCompile(`https://spotify.link/([A-Za-z0-9]+)`)
	ErrInvalidInput         = errors.New("invalid user input for Spotify URI")
	ErrExplicitSong         = errors.New("user does not allow adding explicit songs to the queue")
	ErrSongTooLong          = errors.New("song is too long")
)

// ensure struct implements queue.Publisher
var _ queue.Publisher = (*SpotifyPlayerQueue)(nil)

// TODO: Publisher is an unnecessary struct because there is no state that the publisher tracks here.
type SpotifyPlayerQueue struct {
	OpaqueLinkResolver func(string) string
}

func httpRequestResolver(s string) string {
	if resp, err := http.Get(s); err == nil { //nolint: gosec,noctx
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return ""
		}

		return string(bytes)
	}

	return ""
}

func NewSpotifyPlayerQueue() *SpotifyPlayerQueue {
	return &SpotifyPlayerQueue{
		OpaqueLinkResolver: httpRequestResolver,
	}
}

// Publish will validate that the input matches a valid Spotify URL scheme,
// and then attempt to queue it in the user's Spotify player.
func (s *SpotifyPlayerQueue) Publish(client queue.Queuer, input string, pref *preferences.Preference) (spotify.ID, error) {
	id := parseSpotifyTrackID(input, s.OpaqueLinkResolver)
	if id == "" {
		var err error
		id, err = Search(client, input, pref)
		if err != nil {
			return "", err
		}
	}

	sID := spotify.ID(id)

	if err := ShouldQueue(client, sID, pref); err != nil {
		return sID, err
	}

	return sID, client.QueueSong(context.Background(), sID)
}

func Search(client queue.Queuer, input string, pref *preferences.Preference) (string, error) {
	res, err := client.Search(context.Background(), input, spotify.SearchTypeTrack)
	if err != nil {
		return "", err
	}
	if len(res.Tracks.Tracks) < 1 {
		return "", ErrInvalidInput // no search results found, so most likely a malformed original input
	}
	return res.Tracks.Tracks[0].ID.String(), nil
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
func parseSpotifyTrackID(s string, resolver func(string) string) string {
	// need to check for the opaque link first
	if opaqueSpotifyURLPattern.MatchString(s) {
		s = resolver(s)
	}

	groups := openSpotifyURLPattern.FindStringSubmatch(s)
	if len(groups) < 1 {
		return ""
	}

	return groups[len(groups)-1]
}
