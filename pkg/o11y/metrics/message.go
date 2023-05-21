package metrics

import "time"

type Message struct {
	CreatedAt     *time.Time `column:"created_at"`
	Success       int        `column:"success"` // 0 = failure, 1 = success
	BroadcasterID string     `column:"broadcaster_id"`
	SpotifyTrack  string     `column:"spotify_track"`
}
