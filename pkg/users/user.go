package users

import "time"

type User struct {
	TwitchID            string
	TwitchAccessToken   string
	TwitchRefreshToken  string
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	SpotifyExpiry       *time.Time
}
