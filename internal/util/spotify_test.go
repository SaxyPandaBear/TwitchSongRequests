package util_test

import (
	"testing"

	"github.com/saxypandabear/twitchsongrequests/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/zmb3/spotify/v2"
)

func TestParseTrackDataTruncates(t *testing.T) {
	tracks := []spotify.FullTrack{
		{
			SimpleTrack: spotify.SimpleTrack{
				Name: "foo",
				Artists: []spotify.SimpleArtist{
					{
						Name: "Name1",
					},
					{
						Name: "Other Artist",
					},
				},
				URI: spotify.URI("abc/def"),
			},
			Album: spotify.SimpleAlbum{
				Name: "Some Album",
			},
		},
		{
			SimpleTrack: spotify.SimpleTrack{
				Name: "bar", // order matters for truncation
				Artists: []spotify.SimpleArtist{
					{
						Name: "Name1",
					},
					{
						Name: "Other Artist",
					},
				},
				URI: spotify.URI("abc/def"),
			},
			Album: spotify.SimpleAlbum{
				Name: "Some Album",
			},
		},
	}

	response := util.ParseTrackData(tracks, 1)
	assert.Greater(t, len(tracks), len(response))
	assert.Len(t, response, 1)
	assert.Equal(t, "foo", response[0].Title)
}

func TestParseTrackDataEmpty(t *testing.T) {
	assert.Empty(t, util.ParseTrackData([]spotify.FullTrack{}, 1))
}

func TestSpotifyTrackToPageData(t *testing.T) {
	track := spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{
			Name: "foo",
			Artists: []spotify.SimpleArtist{
				{
					Name: "Name1",
				},
				{
					Name: "Other Artist",
				},
			},
			URI: spotify.URI("abc/def"),
		},
		Album: spotify.SimpleAlbum{
			Name: "Some Album",
		},
	}

	resp := util.SpotifyTrackToPageData(&track)
	assert.NotNil(t, resp)
	assert.Equal(t, "foo", resp.Title)
	assert.Equal(t, "abc/def", resp.URI)
	assert.Equal(t, "Some Album", resp.Album)
	assert.Equal(t, "Name1, Other Artist", resp.Artist)
}
