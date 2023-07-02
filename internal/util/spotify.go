package util

import (
	"strings"

	"github.com/zmb3/spotify/v2"
)

type Track struct {
	Position int
	Title    string
	Artist   string
	Album    string
}

func ParseTrackData(tracks []spotify.FullTrack, limit int) []Track {
	response := make([]Track, 0, limit) // TODO: not sure how many queued songs Spotify will respond with
	for i, tr := range tracks {
		if i >= limit {
			break
		}
		response = append(response, *SpotifyTrackToPageData(&tr, i+1))
	}

	return response
}

func SpotifyTrackToPageData(tr *spotify.FullTrack, pos int) *Track {
	// need to concatenate the list of artist names since they are all separated
	artistNames := make([]string, 0, len(tr.Artists))
	for _, a := range tr.Artists {
		artistNames = append(artistNames, a.Name)
	}

	return &Track{
		Position: pos,
		Title:    tr.Name,
		Artist:   strings.Join(artistNames, ", "),
		Album:    tr.Album.Name,
	}
}
