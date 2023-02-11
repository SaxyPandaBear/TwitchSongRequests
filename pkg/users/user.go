package users

type User struct {
	TwitchID    string
	TwitchAuth  Auth
	SpotifyAuth Auth
}

type Auth struct {
	AccessToken, RefreshToken string
}
