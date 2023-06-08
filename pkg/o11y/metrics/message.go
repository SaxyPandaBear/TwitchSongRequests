package metrics

import "time"

type Message struct {
	CreatedAt     *time.Time `json:"created_at" column:"created_at"`
	Success       int        `column:"success"` // 0 = failure, 1 = success
	BroadcasterID string     `json:"broadcaster_id" column:"broadcaster_id"`
	SpotifyTrack  string     `json:"spotify_track" column:"spotify_track"`
}
