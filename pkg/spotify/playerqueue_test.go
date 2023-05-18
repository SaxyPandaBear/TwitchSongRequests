package spotify

import (
	"testing"
	"time"

	"github.com/saxypandabear/twitchsongrequests/internal/testutil"
	"github.com/saxypandabear/twitchsongrequests/pkg/preferences"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"
)

var testSpotifyTrackURL = "https://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK?si=a99029531fa04a00"

func TestParseSpotifyURL(t *testing.T) {
	tests := map[string]string{
		"https://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK?si=a99029531fa04a00": "3cfOd4CMv2snFaKAnMdnvK",
		"":    "",
		"abc": "",
		"http://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK":                              "",
		"https://open.spotify.com/track/?si=a99029531fa04a00":                               "",
		"https://open.spotify.com/intl-de/track/5Sk39LuvdwuvL84jD01Dum?si=f515f3232b994c5b": "5Sk39LuvdwuvL84jD01Dum",
	}
	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			assert.Equal(t, expected, parseSpotifyTrackID(input))
		})
	}
}

func TestPublish(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages:     make([]spotify.ID, 0, 1),
		GetTrackFunc: testutil.DefaultMockQueuerGetTrackFunc,
	}
	pref := preferences.Preference{
		ExplicitSongs: false,
	}

	err := s.Publish(&q, testSpotifyTrackURL, &pref)
	assert.NoError(t, err)
	assert.Len(t, q.Messages, 1)
	assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", q.Messages[0].String())
}

func TestPublishInvalidInput(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages:     make([]spotify.ID, 0, 1),
		GetTrackFunc: testutil.DefaultMockQueuerGetTrackFunc,
	}
	pref := preferences.Preference{
		ExplicitSongs: false,
	}

	err := s.Publish(&q, "foo", &pref)
	assert.ErrorIs(t, err, ErrInvalidInput)
	assert.Empty(t, q.Messages)
}

func TestPublishFails(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages:     make([]spotify.ID, 0, 1),
		GetTrackFunc: testutil.DefaultMockQueuerGetTrackFunc,
		ShouldFail:   true,
	}
	pref := preferences.Preference{
		ExplicitSongs: false,
	}

	err := s.Publish(&q, "abc123", &pref)
	assert.Error(t, err)
	assert.Empty(t, q.Messages)
}

func TestPublishExplicitSongNotAllowed(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages: make([]spotify.ID, 0, 1),
		GetTrackFunc: func(i spotify.ID) (*spotify.FullTrack, error) {
			track := spotify.FullTrack{}
			track.Explicit = true
			return &track, nil
		},
	}
	pref := preferences.Preference{
		ExplicitSongs: false,
	}

	err := s.Publish(&q, testSpotifyTrackURL, &pref)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrExplicitSong)
	assert.Empty(t, q.Messages)
}

func TestPublishExplicitSongAllowed(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages:     make([]spotify.ID, 0, 2),
		Explicit:     true,
		GetTrackFunc: testutil.DefaultMockQueuerGetTrackFunc,
	}
	pref := preferences.Preference{
		ExplicitSongs: true,
	}

	// not explicit songs should work
	err := s.Publish(&q, testSpotifyTrackURL, &pref)
	assert.NoError(t, err)
	assert.Len(t, q.Messages, 1)
	assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", q.Messages[0].String())

	q.GetTrackFunc = func(_ spotify.ID) (*spotify.FullTrack, error) {
		track := spotify.FullTrack{}
		track.Explicit = true
		return &track, nil
	}

	err = s.Publish(&q, testSpotifyTrackURL, &pref)
	assert.NoError(t, err)
	assert.Len(t, q.Messages, 2)
	for _, id := range q.Messages {
		assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", id.String())
	}
}

func TestShouldQueueExplicitSongs(t *testing.T) {
	q := testutil.MockQueuer{
		Messages: make([]spotify.ID, 0, 1),
		GetTrackFunc: func(i spotify.ID) (*spotify.FullTrack, error) {
			track := spotify.FullTrack{}
			track.Explicit = true
			return &track, nil
		},
	}

	p := preferences.Preference{
		ExplicitSongs: true,
	}

	assert.NoError(t, ShouldQueue(&q, spotify.ID("abc123"), &p))
	assert.Error(t, ShouldQueue(&q, spotify.ID("abc123"), nil))
	p.ExplicitSongs = false
	assert.Error(t, ShouldQueue(&q, spotify.ID("bcd234"), &p))
	assert.Error(t, ShouldQueue(&q, spotify.ID("bcd234"), nil))

	// default does not return a song tagged as explicit
	q.GetTrackFunc = testutil.DefaultMockQueuerGetTrackFunc
	p.ExplicitSongs = true
	assert.NoError(t, ShouldQueue(&q, spotify.ID("abc123"), &p))
	assert.NoError(t, ShouldQueue(&q, spotify.ID("abc123"), nil))
	p.ExplicitSongs = false
	assert.NoError(t, ShouldQueue(&q, spotify.ID("bcd234"), &p))
	assert.NoError(t, ShouldQueue(&q, spotify.ID("bcd234"), nil))
}

func TestShouldQueueMaxSongLength(t *testing.T) {
	q := testutil.MockQueuer{
		Messages: make([]spotify.ID, 0, 1),
		GetTrackFunc: func(i spotify.ID) (*spotify.FullTrack, error) {
			track := spotify.FullTrack{}
			track.Duration = int(time.Hour.Milliseconds())
			return &track, nil
		},
	}

	p := preferences.Preference{}
	assert.NoError(t, ShouldQueue(&q, spotify.ID("abc123"), &p))
	p.MaxSongLength = 1000
	assert.Error(t, ShouldQueue(&q, spotify.ID("bcd234"), &p))

	q.GetTrackFunc = func(i spotify.ID) (*spotify.FullTrack, error) {
		track := spotify.FullTrack{}
		track.Duration = 10
		return &track, nil
	}
	assert.NoError(t, ShouldQueue(&q, spotify.ID("cde345"), &p))
}
