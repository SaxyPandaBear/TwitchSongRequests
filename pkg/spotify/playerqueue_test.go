package spotify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	// assert.Equal(t, "3cfOd4CMv2snFaKAnMdnvK", parseSpotifyTrackID(url))
}
