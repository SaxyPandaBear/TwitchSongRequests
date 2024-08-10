package users

import (
	"time"
)

type User struct {
	TwitchID            string     `column:"id"`
	TwitchAccessToken   string     `column:"twitch_access"`
	TwitchRefreshToken  string     `column:"twitch_refresh"`
	SpotifyAccessToken  string     `column:"spotify_access"`
	SpotifyRefreshToken string     `column:"spotify_refresh"`
	SpotifyExpiry       *time.Time `column:"spotify_expiry"`
	Subscribed          bool       `column:"subscribed"`
	SubscriptionID      string     `column:"subscription_id"`
	Email               string     `column:"email"`
}

func (u *User) IsAuthenticated() bool {
	tAuthed := len(u.TwitchAccessToken) > 0
	sAuthed := len(u.SpotifyAccessToken) > 0
	return tAuthed && sAuthed
}
