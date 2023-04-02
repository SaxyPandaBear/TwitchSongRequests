package spotify

import (
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"
)

var testSpotifyTrackURL = "https://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK?si=a99029531fa04a00"

func TestParseSpotifyURL(t *testing.T) {
	tests := map[string]string{
		"https://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK?si=a99029531fa04a00": "3cfOd4CMv2snFaKAnMdnvK",
		"":    "",
		"abc": "",
		"http://open.spotify.com/track/3cfOd4CMv2snFaKAnMdnvK": "",
		"https://open.spotify.com/track/?si=a99029531fa04a00":  "",
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

	err := s.Publish(&q, testSpotifyTrackURL, false)
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

	err := s.Publish(&q, "foo", false)
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

	err := s.Publish(&q, "abc123", false)
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

	err := s.Publish(&q, "abc123", false)
	assert.Error(t, err)
	assert.Empty(t, q.Messages)
}

func TestPublishExplicitSongAllowed(t *testing.T) {
	s := SpotifyPlayerQueue{}
	q := testutil.MockQueuer{
		Messages:     make([]spotify.ID, 0, 2),
		Explicit:     true,
		GetTrackFunc: testutil.DefaultMockQueuerGetTrackFunc,
	}

	// not explicit songs should work
	err := s.Publish(&q, testSpotifyTrackURL, false)
	assert.NoError(t, err)
	assert.Len(t, q.Messages, 1)
	assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", q.Messages[0].String())

	q.GetTrackFunc = func(_ spotify.ID) (*spotify.FullTrack, error) {
		track := spotify.FullTrack{}
		track.Explicit = true
		return &track, nil
	}

	err = s.Publish(&q, testSpotifyTrackURL, false)
	assert.NoError(t, err)
	assert.Len(t, q.Messages, 2)
	for _, id := range q.Messages {
		assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", id.String())
	}
}
