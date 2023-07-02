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
			},
			Album: spotify.SimpleAlbum{
				Name: "Some Album",
			},
		},
		{
			SimpleTrack: spotify.SimpleTrack{
				Name: "bar",
				Artists: []spotify.SimpleArtist{
					{
						Name: "Name1",
					},
					{
						Name: "Other Artist",
					},
				},
			},
			Album: spotify.SimpleAlbum{
				Name: "Some Album",
			},
		},
		{
			SimpleTrack: spotify.SimpleTrack{
				Name: "baz", // order matters for truncation
				Artists: []spotify.SimpleArtist{
					{
						Name: "Name1",
					},
					{
						Name: "Other Artist",
					},
				},
			},
			Album: spotify.SimpleAlbum{
				Name: "Some Album",
			},
		},
	}

	response := util.ParseTrackData(tracks, 2)
	assert.Greater(t, len(tracks), len(response))
	assert.Len(t, response, 2)
	assert.Equal(t, "foo", response[0].Title)
	assert.Equal(t, "bar", response[1].Title)
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
		},
		Album: spotify.SimpleAlbum{
			Name: "Some Album",
		},
	}

	resp := util.SpotifyTrackToPageData(&track)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.Position) // default value
	assert.Equal(t, "foo", resp.Title)
	assert.Equal(t, "Some Album", resp.Album)
	assert.Equal(t, "Name1, Other Artist", resp.Artist)
}
