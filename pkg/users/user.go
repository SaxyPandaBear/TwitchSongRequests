package users

import "time"

type User struct {
	TwitchID            string
	TwitchAccessToken   string
	TwitchRefreshToken  string
	SpotifyAccessToken  string
	SpotifyRefreshToken string
	SpotifyExpiry       *time.Time
	Subscribed          bool
}

func (u *User) IsAuthenticated() bool {
	tAuthed := len(u.TwitchAccessToken) > 0 && len(u.TwitchRefreshToken) > 0
	sAuthed := len(u.SpotifyAccessToken) > 0 && len(u.SpotifyRefreshToken) > 0
	return tAuthed && sAuthed
}
